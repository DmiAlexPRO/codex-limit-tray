package ui

import (
	"errors"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TestLanguageSelection(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	s := settings.Default()
	path := filepath.Join(t.TempDir(), "settings.json")
	rejectSave := false
	save := func(next settings.Settings) error {
		if rejectSave {
			return errors.New("write failed")
		}
		return next.Save(path)
	}
	v := NewView(a, w, s, func() {}, func() {}, save)
	v.ConfigureUpdates("3.4.1", true, func() {}, func() {})
	retryAt := time.Date(2026, time.November, 12, 10, 30, 0, 0, time.Local)
	releasePage := &url.URL{Scheme: "https", Host: "github.com", Path: "/org/repo/releases/latest"}
	v.SetUpdateState(UpdateState{Stage: UpdateError, RetryAt: retryAt, ReleasePage: releasePage})
	v.UpdateLocalized(nil, func(language string) (string, string) {
		return i18n.Text(language, "connecting"), i18n.Text(language, "waitingReply")
	})
	v.ShowSettings()
	w.Show()
	control := findForm(v.settingsPage.Selected().Content).Items[0].Widget.(*widget.Select)
	if control.Selected != "English" || len(control.Options) != len(i18n.Languages()) {
		t.Fatal("incorrect initial language selector")
	}
	test.Tap(control)
	popup, ok := w.Canvas().Focused().(*widget.PopUpMenu)
	if !ok {
		t.Fatal("language list did not open with keyboard support")
	}
	popup.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	popup.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	saved, err := settings.Load(path, "en")
	if err != nil || saved.Language != "hy" || v.language != "hy" {
		t.Fatalf("selection was not applied and saved: %+v, %v", saved, err)
	}
	tabs := findTabs(w.Content())
	if tabs == nil || tabs.Selected().Text != i18n.Text("hy", "appearance") || tabs.Items[2].Text != i18n.Text("hy", "updates") {
		t.Fatal("visible settings were not translated immediately")
	}
	if v.status != i18n.Text("hy", "connecting") || v.diagnostic.Text != i18n.Text("hy", "waitingReply") {
		t.Fatal("status was not translated")
	}
	if !v.updates.state.RetryAt.Equal(retryAt) || v.updates.state.ReleasePage != releasePage ||
		!v.updates.openRelease.Visible() || v.updates.openRelease.Text != i18n.Text("hy", "openReleasePage") {
		t.Fatal("rate limit or localized release link was lost on language change")
	}
	rejectSave = true
	control = findForm(tabs.Selected().Content).Items[0].Widget.(*widget.Select)
	control.SetSelected("Русский")
	if control.Selected != "Հայերեն" || v.settings.Language != "hy" || v.language != "hy" {
		t.Fatal("failed write changed language")
	}
	w.Close()

	after := a.NewWindow("Test after restart")
	restarted := NewView(a, after, saved, func() {}, func() {}, save)
	if restarted.settingsPage.Items[0].Text != "Տեսք" {
		t.Fatal("saved language not applied on restart")
	}
	after.Close()
}

func TestLanguagePagesRender(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	for _, lang := range i18n.Languages() {
		t.Run(lang.Code, func(t *testing.T) {
			s := settings.Default()
			w := a.NewWindow("Languages")
			defer w.Close()
			v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
			v.ConfigureUpdates("3.1.0", true, func() {}, func() {})
			v.SetUpdateState(UpdateState{Stage: UpdateAvailable, Version: "3.2.0", CanInstall: true})
			v.Update([]quota.Window{
				{Remaining: 57, Minutes: 300, Reset: time.Date(2026, 11, 14, 14, 25, 0, 0, time.UTC)},
				{Remaining: 82, Minutes: 10080},
			}, i18n.Text(lang.Code, "updated", "14:25"), i18n.Text(lang.Code, "requestOK"))
			w.Resize(fyne.NewSize(420, 600))
			w.Show()
			v.ShowSettings()
			findForm(v.settingsPage.Selected().Content).Items[0].Widget.(*widget.Select).SetSelected(lang.Name)
			if v.language != lang.Code {
				t.Fatalf("selection produced locale %q", v.language)
			}
			v.showMain()
			checkButtons(t, v.mainPage)
			v.ShowSettings()
			for page := range v.settingsPage.Items {
				v.settingsPage.SelectIndex(page)
				img := w.Canvas().Capture()
				if img.Bounds().Empty() {
					t.Fatalf("page %d did not render", page)
				}
				scroll := v.settingsPage.Selected().Content.(*container.Scroll)
				content := scroll.Content.(*fyne.Container).Objects[0]
				checkButtons(t, content)
				if content.Size().Width > scroll.Size().Width-theme.ScrollBarSize()+0.1 {
					t.Errorf("page %d overflows the scrollbar gutter: content %.1f, viewport %.1f", page, content.Size().Width, scroll.Size().Width)
				}
				if form := findForm(content); form != nil {
					for _, item := range form.Items {
						if item.Widget.Size().Width < item.Widget.MinSize().Width {
							t.Errorf("form control %q is too narrow", item.Text)
						}
					}
				}
				if page != 0 || os.Getenv("CODEX_LIMIT_PREVIEW") != "1" {
					continue
				}
				dir := filepath.Join("..", "..", "dist", "fyne-preview")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(filepath.Join(dir, "language-"+lang.Code+".png"))
				if err != nil {
					t.Fatal(err)
				}
				err = errors.Join(png.Encode(f, img), f.Close())
				if err != nil {
					t.Fatal(err)
				}
			}
			checkLanguageDialogs(t, v, w, lang.Code)
		})
	}
}

func checkLanguageDialogs(t *testing.T, v *View, w fyne.Window, language string) {
	t.Helper()
	tr := func(key string) string { return i18n.Text(language, key) }
	v.settingsPage.SelectIndex(0)
	root := v.settingsPage.Selected().Content.(*container.Scroll).Content
	for _, key := range []string{"accent", "lowColor"} {
		button := resetDialogButton(root, tr(key)+"…")
		if button == nil {
			t.Fatalf("missing color action for %s", key)
		}
		test.Tap(button)
		overlay := w.Canvas().Overlays().Top()
		confirm := resetDialogButton(overlay, tr("confirm"))
		if confirm == nil {
			t.Fatalf("color dialog for %s is not localized", key)
		}
		checkButtons(t, confirm)
		cancel := resetDialogButton(overlay, tr("earnedResetCancel"))
		if cancel == nil {
			t.Fatal("missing localized color cancel action")
		}
		checkButtons(t, cancel)
		test.Tap(cancel)
	}

	v.settingsPage.SelectIndex(1)
	content := v.settingsPage.Selected().Content.(*container.Scroll).Content.(*fyne.Container).Objects[0].(*fyne.Container)
	path := content.Objects[2].(*widget.Entry)
	test.TapSecondary(path)
	menu, ok := w.Canvas().Focused().(*widget.PopUpMenu)
	if !ok {
		t.Fatal("path context menu did not open")
	}
	markup := test.RenderToMarkup(w.Canvas())
	for _, key := range []string{"cut", "copy", "paste", "selectAll"} {
		if !strings.Contains(markup, tr(key)) {
			t.Errorf("path context menu is missing %s", key)
		}
	}
	menu.Hide()

	v.showMain()
	v.ConfigureReset(func() { t.Error("cancelled reset was requested") })
	v.SetResetCredits(&quota.ResetCredits{Available: 1}, "")
	test.Tap(v.resets.button)
	overlay := w.Canvas().Overlays().Top()
	if overlay == nil || resetDialogButton(overlay, tr("earnedResetConfirm")) == nil {
		t.Fatal("reset confirmation is not localized")
	}
	cancel := resetDialogButton(overlay, tr("earnedResetCancel"))
	if cancel == nil {
		t.Fatal("missing localized reset cancel action")
	}
	test.Tap(cancel)
}

func TestWidgetLanguageFollowsSelection(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Widgets")
	defer w.Close()
	v := NewView(a, w, settings.Default(), func() {}, func() {}, func(settings.Settings) error { return nil })
	v.ShowSettings()
	w.Show()
	for _, locale := range []struct {
		name, advanced, paste string
	}{{"Čeština", "Pokročilé", "Vložit"}, {"한국어", "고급", "붙여넣기"}, {"English", "Advanced", "Paste"}} {
		control := findForm(v.settingsPage.Selected().Content).Items[0].Widget.(*widget.Select)
		control.SetSelected(locale.name)
		if lang.L("Advanced") != locale.advanced || lang.L("Paste") != locale.paste {
			t.Errorf("Fyne strings did not change to %s", locale.name)
		}
	}
}
