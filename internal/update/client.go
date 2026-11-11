// Package update reads release metadata and downloads verified Windows installers.
// It never executes downloaded code or changes the installed application.
package update

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Set by the official build with -ldflags -X. Development builds have no
// default repository or trust key and must not contact a guessed publisher.
var (
	Version    = "3.0.0-preview"
	Repository string
	PublicKey  string
)

const (
	maxMetadata  = 4 << 20
	maxInstaller = 128 << 20
	maxNotes     = 4096
)

type Release struct {
	Version          string
	Notes            string
	tag, repository  string
	newer, hasAssets bool
}

func (r *Release) IsNewer() bool      { return r != nil && r.newer }
func (r *Release) HasInstaller() bool { return r != nil && r.hasAssets }

func InstallKeyConfigured() bool {
	_, err := decodeKey(PublicKey)
	return err == nil
}

type client struct {
	repo, current, publicKey string
	http                     *http.Client
}

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}/[A-Za-z0-9_.-]{1,100}$`)

func (c client) validate() (version, error) {
	if !repoPattern.MatchString(c.repo) || strings.HasSuffix(c.repo, "/.") || strings.HasSuffix(c.repo, "/..") {
		return version{}, errors.New("update repository not configured")
	}
	v, ok := parseVersion(c.current)
	if !ok {
		return version{}, errors.New("invalid app version")
	}
	return v, nil
}

func redirect(req *http.Request, via []*http.Request) error {
	u := req.URL
	if len(via) == 0 || len(via) >= 4 || via[0].URL.Host != "github.com" || u.Scheme != "https" || u.User != nil {
		return errors.New("redirect not allowed")
	}
	switch u.Host {
	case "release-assets.githubusercontent.com", "objects.githubusercontent.com":
		return nil
	default:
		return errors.New("unknown download server")
	}
}

func configured() client {
	return client{repo: Repository, current: Version, publicKey: PublicKey, http: &http.Client{CheckRedirect: redirect}}
}

// Check returns a newer release or the exact current release with its notes.
// A stable build only follows GitHub's latest stable release; a preview picks
// the highest compatible SemVer among the 30 most recently published entries.
func Check(ctx context.Context) (*Release, error) { return configured().check(ctx) }

type asset struct {
	Name string `json:"name"`
}
type githubRelease struct {
	Tag        string  `json:"tag_name"`
	Body       string  `json:"body"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
}

func (r githubRelease) packaged() bool {
	hasInstaller, hasSignature := false, false
	for _, a := range r.Assets {
		switch a.Name {
		case AssetName:
			hasInstaller = true
		case AssetName + ".sig":
			hasSignature = true
		}
	}
	return hasInstaller && hasSignature
}

func plainNotes(body string) string {
	var b strings.Builder
	count := 0
	for _, r := range body {
		if r != '\n' && r != '\t' && (unicode.IsControl(r) || unicode.In(r, unicode.Cf)) {
			continue
		}
		if count == maxNotes {
			b.WriteRune('…')
			break
		}
		b.WriteRune(r)
		count++
	}
	return strings.TrimSpace(b.String())
}

func (c client) check(ctx context.Context) (*Release, error) {
	current, err := c.validate()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	endpoint := "https://api.github.com/repos/" + c.repo + "/releases"
	stable := len(current.pre) == 0
	if stable {
		endpoint += "/latest"
	} else {
		endpoint += "?per_page=30"
	}
	data, err := c.bytes(ctx, endpoint, maxMetadata)
	// /latest returns 404 if a public repository has no stable release yet.
	// The list endpoint returns [] for an empty repo: its 404 is a real error.
	if stable && errors.Is(err, errNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var releases []githubRelease
	if stable {
		releases = make([]githubRelease, 1)
		err = json.Unmarshal(data, &releases[0])
	} else {
		err = json.Unmarshal(data, &releases)
	}
	if err != nil || releases == nil || (stable && releases[0].Tag == "") {
		return nil, errors.New("invalid GitHub release data")
	}
	var best *Release
	bestVersion := current
	for _, r := range releases {
		name := strings.TrimPrefix(r.Tag, "v")
		v, ok := parseVersion(name)
		if !ok || r.Draft || (stable && (r.Prerelease || len(v.pre) != 0)) {
			continue
		}
		order := v.compare(current)
		if order < 0 || (order == 0 && name != c.current) {
			continue
		}
		if best != nil && v.compare(bestVersion) <= 0 {
			continue
		}
		bestVersion = v
		best = &Release{Version: name, Notes: plainNotes(r.Body), tag: r.Tag, repository: c.repo, newer: order > 0, hasAssets: r.packaged()}
	}
	return best, nil
}

var errNotFound = errors.New("file or release not found on GitHub")

func (c client) get(ctx context.Context, endpoint string, limit int64) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.New("invalid update URL")
	}
	req.Header.Set("User-Agent", "CodexLimitTray-Updater")
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("could not contact the update server")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, errNotFound
		}
		return nil, fmt.Errorf("update server returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		resp.Body.Close()
		return nil, errors.New("response exceeds the size limit")
	}
	return resp, nil
}

func (c client) bytes(ctx context.Context, endpoint string, limit int64) ([]byte, error) {
	resp, err := c.get(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("response exceeds the size limit")
	}
	return data, nil
}

// Download returns a unique verified temporary EXE. The caller owns it and
// must remove it after the installer exits. Partial downloads are never returned.
func Download(ctx context.Context, r *Release) (string, error) {
	return configured().download(ctx, r, false)
}

// DownloadForInstall also allows the exact current version for explicit first
// installation of a portable copy. Never use this function for automatic updates.
func DownloadForInstall(ctx context.Context, r *Release) (string, error) {
	return configured().download(ctx, r, true)
}

func (c client) download(ctx context.Context, r *Release, allowCurrent bool) (string, error) {
	current, err := c.validate()
	if err != nil {
		return "", err
	}
	key, err := decodeKey(c.publicKey)
	if err != nil {
		return "", err
	}
	if r == nil || r.repository != c.repo || strings.TrimPrefix(r.tag, "v") != r.Version {
		return "", errors.New("check the release again")
	}
	v, ok := parseVersion(r.Version)
	if !ok || (v.compare(current) <= 0 && !(allowCurrent && r.Version == c.current)) || (len(current.pre) == 0 && len(v.pre) != 0) {
		return "", errors.New("incompatible update version")
	}
	if !r.hasAssets {
		return "", errors.New("release has no signed Windows x64 installer")
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()
	// Never follow a browser_download_url or API asset URL from remote JSON.
	endpoint := "https://github.com/" + c.repo + "/releases/download/" + r.tag + "/" + AssetName
	data, err := c.bytes(ctx, endpoint+".sig", 256)
	if err != nil {
		return "", err
	}
	sig, err := decodeSignature(data)
	if err != nil {
		return "", err
	}
	resp, err := c.get(ctx, endpoint, maxInstaller)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	dir, f, name, err := newInstaller()
	if err != nil {
		return "", err
	}
	defer dir.Close()
	verified := false
	defer func() {
		if !verified {
			_ = dir.Remove(name)
		}
	}()
	hash := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(f, hash), io.LimitReader(resp.Body, maxInstaller+1))
	closeErr := f.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if n == 0 || n > maxInstaller || (resp.ContentLength >= 0 && n != resp.ContentLength) {
		return "", errors.New("invalid installer size")
	}
	if !ed25519.Verify(key, SignatureMessage(r.Version, hex.EncodeToString(hash.Sum(nil))), sig) {
		return "", errors.New("invalid update signature; installation canceled")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	verified = true
	return f.Name(), nil
}
