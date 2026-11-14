package update

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type transport func(*http.Request) (*http.Response, error)

func (rt transport) RoundTrip(req *http.Request) (*http.Response, error) { return rt(req) }

func reply(body string, code int) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), ContentLength: -1, Header: http.Header{}}
}

func mock(current string, rt transport) client {
	return client{repo: "org/repo", current: current, http: &http.Client{Transport: rt, CheckRedirect: redirect}}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		current, metadata, want string
		newer                   bool
	}{
		{"1.2.0", `{"tag_name":"v1.10.0","body":"Notes\u202E" }`, "1.10.0", true},
		{"1.2.0", `{"tag_name":"v1.2.0","body":"Notes"}`, "1.2.0", false},
		{"1.2.0+ours", `{"tag_name":"v1.2.0+other"}`, "", false},
		{"1.2.0", `{"tag_name":"v0.9.0"}`, "", false},
		{"1.2.0", `{"tag_name":"v4.0.0","draft":true}`, "", false},
		{"1.2.0", `{"tag_name":"v4.0.0","prerelease":true}`, "", false},
		{"1.2.0", `{"tag_name":"v4.0.0-rc.1"}`, "", false},
		{"1.2.0-preview.1", `[{"tag_name":"v2.0.0","draft":true},{"tag_name":"v../bad"},{"tag_name":"v1.2.0-preview.11","prerelease":true,"body":"Notes"},{"tag_name":"v1.2.0-preview.2"}]`, "1.2.0-preview.11", true},
		{"1.2.0-preview", `[{"tag_name":"v1.2.0","body":"Notes"},{"tag_name":"v1.2.0-rc.1"}]`, "1.2.0", true},
	}
	for _, tc := range tests {
		c := mock(tc.current, func(req *http.Request) (*http.Response, error) {
			if req.URL.Host != "api.github.com" || req.URL.Scheme != "https" {
				t.Fatalf("bad URL: %v", req.URL)
			}
			want := "/repos/org/repo/releases/latest"
			if strings.Contains(tc.current, "preview") {
				want = "/repos/org/repo/releases?per_page=30"
			}
			if req.URL.RequestURI() != want {
				t.Fatalf("wrong endpoint %s", req.URL)
			}
			return reply(tc.metadata, 200), nil
		})
		r, err := c.check(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if tc.want == "" && r == nil {
			continue
		}
		if r == nil || r.Version != tc.want || r.IsNewer() != tc.newer || r.Notes != "Notes" {
			t.Fatalf("%s + %s: got %+v", tc.current, tc.metadata, r)
		}
	}
}

func TestSafetyLimits(t *testing.T) {
	for _, response := range []*http.Response{
		reply("not json", 200), reply(strings.Repeat(" ", maxMetadata+1), 200), reply("", 429),
		{StatusCode: 200, ContentLength: maxMetadata + 100, Body: io.NopCloser(strings.NewReader(""))},
	} {
		c := mock("1.0.0", func(*http.Request) (*http.Response, error) { return response, nil })
		if _, err := c.check(context.Background()); err == nil {
			t.Fatal("accepted invalid response")
		}
	}
	if n := len([]rune(plainNotes(strings.Repeat("я", 9000)))); n != maxNotes+1 {
		t.Fatal(n)
	}
	for _, repo := range []string{"", "org/../other", "org/..", "evil.com:443/foo", "a/b?token=1"} {
		c := client{repo: repo, current: "1.0.0"}
		if _, err := c.check(context.Background()); err == nil {
			t.Fatal(repo)
		}
	}
	c := mock("1.0.0", func(req *http.Request) (*http.Response, error) { return nil, req.Context().Err() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.check(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestGitHubRateLimit(t *testing.T) {
	reset := time.Now().Add(10 * time.Minute).Truncate(time.Second)
	for _, tc := range []struct {
		name   string
		code   int
		header string
		value  string
		limit  bool
	}{
		{"primary", 403, "X-RateLimit-Remaining", "0", true},
		{"secondary", 403, "Retry-After", "30", true},
		{"too many requests", 429, "", "", true},
		{"unrelated forbidden", 403, "X-RateLimit-Remaining", "4", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock("1.0.0", func(*http.Request) (*http.Response, error) {
				resp := reply("untrusted server text", tc.code)
				resp.Header.Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
				if tc.header != "" {
					resp.Header.Set(tc.header, tc.value)
				}
				return resp, nil
			})
			_, err := c.check(context.Background())
			var limit *RateLimitError
			if errors.As(err, &limit) != tc.limit {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.limit {
				if err == nil || err.Error() != "update server returned HTTP 403" {
					t.Fatalf("ordinary 403 changed: %v", err)
				}
				return
			}
			earliest := time.Now().Add(20 * time.Second)
			validRetry := limit.RetryAt.After(earliest) && limit.RetryAt.Before(reset)
			if tc.name == "primary" {
				validRetry = limit.RetryAt.Equal(reset)
			}
			if !validRetry || limit.ReleasePage.String() != "https://github.com/org/repo/releases/latest" {
				t.Fatalf("unexpected retry time or release URL: %+v", limit)
			}
		})
	}
}

func TestRetryAt(t *testing.T) {
	now := time.Date(2026, time.November, 12, 9, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		retry, reset, remaining string
		want                    time.Duration
	}{
		{"120", strconv.FormatInt(now.Add(time.Hour).Unix(), 10), "5", 2 * time.Minute},
		{"30", strconv.FormatInt(now.Add(time.Minute).Unix(), 10), "0", time.Minute},
		{now.Add(5 * time.Minute).Format(http.TimeFormat), strconv.FormatInt(now.Add(time.Minute).Unix(), 10), "0", 5 * time.Minute},
		{"garbage", strconv.FormatInt(now.Add(-time.Minute).Unix(), 10), "0", time.Minute},
		{"9223372036854775807", "", "", time.Minute},
	} {
		headers := make(http.Header)
		headers.Set("Retry-After", tc.retry)
		headers.Set("X-RateLimit-Reset", tc.reset)
		headers.Set("X-RateLimit-Remaining", tc.remaining)
		if got := retryAt(headers, now); got.Sub(now) != tc.want {
			t.Errorf("retry %q, reset %q: got %v, want %v", tc.retry, tc.reset, got.Sub(now), tc.want)
		}
	}
}

func TestRedirectPolicy(t *testing.T) {
	first, _ := http.NewRequest("GET", "https://github.com/org/repo/releases/download/v1.0.0/"+AssetName, nil)
	for _, s := range []string{"https://github.com/other", "http://objects.githubusercontent.com/a", "https://release-assets.githubusercontent.com.evil/a", "https://evil@objects.githubusercontent.com/a", "https://objects.githubusercontent.com:444/a", "https://127.0.0.1/x"} {
		u, _ := url.Parse(s)
		if redirect(&http.Request{URL: u}, []*http.Request{first}) == nil {
			t.Fatal(s)
		}
	}
	req, _ := http.NewRequest("GET", "https://release-assets.githubusercontent.com/signed?data=test", nil)
	if err := redirect(req, []*http.Request{first}); err != nil {
		t.Fatal(err)
	}
	api, _ := http.NewRequest("GET", "https://api.github.com/repos/org/repo/releases/latest", nil)
	if redirect(req, []*http.Request{api}) == nil {
		t.Fatal("metadata redirected")
	}
	if redirect(req, []*http.Request{first, req, req, req}) == nil {
		t.Fatal("redirect loop")
	}
}

func TestDownload(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	installer := "synthetic installer bytes"
	sum := sha256.Sum256([]byte(installer))
	// Build the envelope independently of SignatureMessage; this is the CI contract.
	sign := func(version string) string {
		msg := "CodexLimitTray update v1\n" + version + "\nCodexLimitTray-windows-x64-Setup.exe\n" + hex.EncodeToString(sum[:]) + "\n"
		return base64.StdEncoding.EncodeToString(ed25519.Sign(priv, []byte(msg))) + "\n"
	}
	const metadata = `{"tag_name":"v1.1.0","assets":[{"name":"CodexLimitTray-windows-x64-Setup.exe","browser_download_url":"https://evil.invalid/file"},{"name":"CodexLimitTray-windows-x64-Setup.exe.sig"}]}`
	cache := setTestCache(t)
	tmp := filepath.Join(cache, "CodexLimitTray", "updates")
	for _, tc := range []struct {
		name, current, payload, signature string
		ok                                bool
	}{
		{"newer", "1.0.0", installer, sign("1.1.0"), true},
		{"tampered", "1.0.0", installer + "?", sign("1.1.0"), false},
		{"relabelled", "1.0.0", installer, sign("1.0.1"), false},
		{"bad signature encoding", "1.0.0", installer, "garbage", false},
		{"same no update", "1.1.0", installer, sign("1.1.0"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock(tc.current, func(req *http.Request) (*http.Response, error) {
				base := "https://github.com/org/repo/releases/download/v1.1.0/" + AssetName
				switch req.URL.String() {
				case "https://api.github.com/repos/org/repo/releases/latest":
					return reply(metadata, 200), nil
				case base:
					return reply(tc.payload, 200), nil
				case base + ".sig":
					return reply(tc.signature, 200), nil
				default:
					t.Fatalf("untrusted URL: %v", req.URL)
					return nil, nil
				}
			})
			c.publicKey = base64.StdEncoding.EncodeToString(pub)
			r, err := c.check(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			path, err := c.download(context.Background(), r)
			if tc.ok != (err == nil) {
				t.Fatalf("path=%s err=%v", path, err)
			}
			if tc.ok {
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, []byte(installer)) {
					t.Fatalf("bad installer: %v", err)
				}
				if err = os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			if files, _ := filepath.Glob(filepath.Join(tmp, "*.exe")); len(files) != 0 {
				t.Fatalf("failed to clean %v", files)
			}
		})
	}
	c := mock("1.0.0", func(*http.Request) (*http.Response, error) {
		t.Fatal("unverified download made a request")
		return nil, nil
	})
	r := &Release{Version: "1.1.0", tag: "v1.1.0", repository: c.repo, hasAssets: true}
	if _, err := c.download(context.Background(), r); err == nil {
		t.Fatal("accepted missing embedded key")
	}
	c.publicKey = base64.StdEncoding.EncodeToString(pub)
	for _, invalid := range []*Release{nil, {Version: "1.1.0"}, {Version: "1.1.0", repository: c.repo, tag: "v2.0.0"}, {Version: "0.9.0", tag: "v0.9.0", repository: c.repo, hasAssets: true}, {Version: "1.1.0", tag: "v1.1.0", repository: c.repo}} {
		if _, err := c.download(context.Background(), invalid); err == nil {
			t.Fatal("accepted invalid release")
		}
	}
}

func TestSemVer(t *testing.T) {
	for _, s := range []string{"1.2", "01.0.0", "1.0.0-01", "v1.2.3", "1.2.3/evil", "1.2.3+", "1.0.0-rc..2", "1.2.3\n"} {
		if _, ok := parseVersion(s); ok {
			t.Fatal(s)
		}
	}
	ascending := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.2.0", "10.0.0", "100000000000000000000000.1.0"}
	for i := 1; i < len(ascending); i++ {
		a, _ := parseVersion(ascending[i-1])
		b, _ := parseVersion(ascending[i])
		if a.compare(b) >= 0 || b.compare(a) <= 0 {
			t.Fatalf("%s vs %s", ascending[i-1], ascending[i])
		}
	}
}
