package ui

import (
	"image/color"
	"net/url"
	"strings"
	"testing"
	"time"

	"codexlimittray/internal/settings"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestUpdatesPage(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("updates")
	s := settings.Default()
	v := NewView(a, w, s, func() {}, func() {}, func(next settings.Settings) error { s = next; return nil })
	checks, installs := 0, 0
	v.ConfigureUpdates("1.0.0", true, func() { checks++ }, func() { installs++ })
	v.ShowUpdates()
	w.Show()
	p := v.updates
	if !v.onSettings || v.settingsPage.Selected() != p.tab || p.install.Visible() || !p.automatic.Checked {
		t.Fatal("unexpected initial update page")
	}
	if p.automatic.MinSize().Width > w.Canvas().Size().Width-24 {
		t.Fatalf("automatic-update option does not fit: %v, canvas=%v", p.automatic.MinSize(), w.Canvas().Size())
	}
	test.Tap(p.check)
	v.SetUpdateState(UpdateState{Stage: UpdateChecking})
	if !p.check.Disabled() || p.automatic.Disabled() {
		t.Fatal("during check, only the automatic-update preference should stay editable")
	}
	test.Tap(p.automatic)
	if checks != 1 || s.AutoUpdate || v.settings.AutoUpdate {
		t.Fatal("manual check or opt-out not forwarded")
	}
	notes := "# Literal ![image](https://example.org/test)"
	v.SetUpdateState(UpdateState{Stage: UpdateAvailable, Version: "1.1.0", Notes: notes, CanInstall: true})
	if !strings.Contains(p.version.Text, "1.1.0") || !strings.Contains(p.notes.Text, notes) || !p.install.Visible() || p.install.Disabled() {
		t.Fatal("release notes or install action missing")
	}
	test.Tap(p.install)
	v.SetUpdateState(UpdateState{Stage: UpdateInstalling})
	if !p.install.Disabled() || installs != 1 || !p.check.Disabled() || p.automatic.Disabled() {
		t.Fatal("manual actions must be locked during installation, not the preference")
	}
	v.SetUpdateState(UpdateState{Stage: UpdateError, Message: "Повторите попытку", CanInstall: true})
	if !strings.Contains(p.status.Text, "Повторите попытку") || p.check.Disabled() || p.install.Disabled() {
		t.Fatal("cannot retry after failure")
	}
	v.SetUpdateState(UpdateState{Stage: UpdateCurrent, Version: "1.0.0", Notes: "Заметки установленной версии"})
	if p.install.Visible() || !p.notes.Visible() || !strings.Contains(p.notes.Text, "Заметки установленной версии") || !p.version.Visible() {
		t.Fatal("release notes for the current version must remain visible without a reinstall action")
	}
	v.SetUpdateState(UpdateState{Stage: UpdateCurrent})
	if p.install.Visible() || p.notes.Visible() || p.version.Visible() {
		t.Fatal("stale release shown")
	}
}

func TestPortableUpdateOpensGitHub(t *testing.T) {
	a := &settingsLinkApp{App: test.NewApp()}
	defer a.Quit()
	w := a.NewWindow("portable")
	s := settings.Default()
	s.Language = "ru"
	v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { t.Fatal("settings modified"); return nil })
	v.ConfigureUpdates("1.0.0", false, func() {}, func() { t.Fatal("portable installer invoked") })
	v.ShowUpdates()
	w.Show()
	p := v.updates
	link := &url.URL{Scheme: "https", Host: "github.com", Path: "/org/repo/releases/tag/v1.1.0"}
	v.SetUpdateState(UpdateState{Stage: UpdateAvailable, Version: "1.1.0", ReleasePage: link, CanInstall: true})
	if !p.automatic.Disabled() || p.install.Visible() || !p.openRelease.Visible() ||
		p.openRelease.Text != "Скачать с GitHub" || a.opened != "" || w.Canvas().Capture().Bounds().Empty() {
		t.Fatal("portable mode must show a GitHub button without opening it automatically")
	}
	test.Tap(p.openRelease)
	if a.opened != link.String() {
		t.Fatalf("opened %q, want %q", a.opened, link)
	}
	v.SetUpdateState(UpdateState{Stage: UpdateCurrent, Version: "1.0.0"})
	if p.openRelease.Visible() || p.install.Visible() {
		t.Fatal("current portable version must not offer a download")
	}
}

func TestRateLimitShowsRetryAndReleasePage(t *testing.T) {
	a := &settingsLinkApp{App: test.NewApp()}
	defer a.Quit()
	w := a.NewWindow("rate limited updates")
	s := settings.Default()
	s.Language = "ru"
	v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
	v.ConfigureUpdates("1.0.0", true, func() {}, func() {})
	v.ShowUpdates()
	w.Show()
	p := v.updates
	if p.openRelease.Visible() {
		t.Fatal("release button visible before rate limit")
	}
	retryAt := time.Date(2026, time.November, 12, 10, 30, 0, 0, time.Local)
	link := &url.URL{Scheme: "https", Host: "github.com", Path: "/org/repo/releases/latest"}
	v.SetUpdateState(UpdateState{Stage: UpdateError, RetryAt: retryAt, ReleasePage: link})
	if !p.openRelease.Visible() || p.openRelease.Text != "Открыть страницу релиза" ||
		!strings.Contains(p.status.Text, "2026-11-12 10:30:00") || strings.Contains(p.status.Text, "403") ||
		p.install.Visible() || a.opened != "" || w.Canvas().Capture().Bounds().Empty() {
		t.Fatalf("invalid rate limit display: %s", p.status.Text)
	}
	test.Tap(p.openRelease)
	if a.opened != link.String() {
		t.Fatalf("opened %q, want %q", a.opened, link)
	}
	v.SetUpdateState(UpdateState{Stage: UpdateChecking})
	if p.openRelease.Visible() {
		t.Fatal("stale release button")
	}
}

func TestButtonTextContrastsWithCustomAccent(t *testing.T) {
	for _, tc := range []struct {
		accent string
		want   color.Color
	}{{"66D9C1", color.Black}, {"000020", color.White}, {"FFFFFF", color.Black}} {
		s := settings.Default()
		s.Accent = tc.accent
		if got := (appTheme{s}).Color(theme.ColorNameForegroundOnPrimary, theme.VariantDark); got != tc.want {
			t.Fatalf("foreground on %s: %v, want %v", tc.accent, got, tc.want)
		}
	}
}
