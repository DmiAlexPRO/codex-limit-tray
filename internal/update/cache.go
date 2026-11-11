package update

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var installerFile = regexp.MustCompile(`^CodexLimitTray-update-[0-9a-f]{32}\.exe$`)

// Open each child relative to the trusted user cache root. Never traverse an
// application-owned path component that has become a symbolic link or junction.
func cacheRoot(create bool) (*os.Root, error) {
	base, err := os.UserCacheDir()
	if err != nil || !filepath.IsAbs(base) {
		return nil, errors.New("user cache directory unavailable")
	}
	parent, err := os.OpenRoot(base)
	if err != nil {
		return nil, err
	}
	for _, sub := range []string{"CodexLimitTray", "updates"} {
		child, err := cacheChild(parent, sub, create)
		parent.Close()
		if err != nil {
			return nil, err
		}
		parent = child
	}
	return parent, nil
}

func cacheChild(parent *os.Root, name string, create bool) (*os.Root, error) {
	if create {
		err := parent.Mkdir(name, 0700)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return nil, err
		}
	}
	info, err := parent.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || reparse(info) {
		return nil, errors.New("unsafe cache directory")
	}
	return parent.OpenRoot(name)
}

func newInstaller() (*os.Root, *os.File, string, error) {
	dir, err := cacheRoot(true)
	if err != nil {
		return nil, nil, "", err
	}
	// Directory and children inherit the per-user LocalAppData ACL on Windows.
	// Do not replace the inherited ACL: restricted user tokens need its ACEs.
	var random [16]byte
	rand.Read(random[:])
	name := "CodexLimitTray-update-" + hex.EncodeToString(random[:]) + ".exe"
	f, err := dir.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		dir.Close()
		return nil, nil, "", err
	}
	return dir, f, name, nil
}

// Cleanup removes only recognized regular installers older than an hour.
// It is safe to call on startup with a missing cache or an in-use installer.
func Cleanup() {
	dir, err := cacheRoot(false)
	if err != nil {
		return
	}
	defer dir.Close()
	d, err := dir.Open(".")
	if err != nil {
		return
	}
	defer d.Close()
	cutoff := time.Now().Add(-time.Hour)
	for {
		names, err := d.Readdirnames(64)
		for _, name := range names {
			if !installerFile.MatchString(name) {
				continue
			}
			st, e := dir.Lstat(name)
			if e == nil && st.Mode().IsRegular() && !reparse(st) && st.ModTime().Before(cutoff) {
				_ = dir.Remove(name)
			}
		}
		if err != nil || len(names) == 0 {
			return
		}
	}
}
