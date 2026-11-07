package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setTestCache(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("LocalAppData", tmp)
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("HOME", tmp)
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cache, 0700); err != nil {
		t.Fatal(err)
	}
	return cache
}

func TestCleanupScopesFiles(t *testing.T) {
	cache := setTestCache(t)
	Cleanup()
	dir := filepath.Join(cache, "CodexLimitTray", "updates")
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cleanup created a directory")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	mk := func(dir, name string, age time.Duration) string {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
		mod := time.Now().Add(-age)
		if err := os.Chtimes(p, mod, mod); err != nil {
			t.Fatal(err)
		}
		return p
	}
	file := func(ch string) string { return "CodexLimitTray-update-" + strings.Repeat(ch, 32) + ".exe" }
	old := mk(dir, file("a"), 2*time.Hour)
	keep := []string{
		mk(dir, file("b"), time.Second),
		mk(dir, "CodexLimitTray-update-not-ours.exe", 2*time.Hour),
		mk(cache, file("c"), 2*time.Hour),
	}
	subdir := filepath.Join(dir, file("d"))
	if err := os.Mkdir(subdir, 0700); err != nil {
		t.Fatal(err)
	}
	keep = append(keep, subdir, mk(subdir, file("e"), 2*time.Hour))
	link := filepath.Join(dir, file("f"))
	if err := os.Symlink(keep[2], link); err == nil {
		keep = append(keep, link)
	}
	Cleanup()
	if _, err := os.Stat(old); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale installer retained: %v", err)
	}
	for _, path := range keep {
		if _, err := os.Lstat(path); err != nil {
			t.Fatal(err)
		}
	}
}
