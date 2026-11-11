package ui

import (
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
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
	control, notice := v.languageSelector()
	w.SetContent(control)
	w.Show()
	if control.Selected != "English" || len(control.Options) != 19 || notice.Visible() {
		t.Fatal("incorrect initial language selector")
	}
	test.Tap(control)
	if w.Canvas().Overlays().Top() == nil {
		t.Fatal("language list did not open")
	}
	w.Canvas().Overlays().Top().Hide()
	control.SetSelected("日本語")
	saved, err := settings.Load(path, "en")
	if err != nil || saved.Language != "ja" || !notice.Visible() || v.language != "en" {
		t.Fatalf("selection was not saved for the next launch: %+v, %v", saved, err)
	}
	rejectSave = true
	control.SetSelected("Русский")
	if control.Selected != "日本語" || v.settings.Language != "ja" {
		t.Fatal("failed write changed language")
	}
	w.Close()

	after := a.NewWindow("Test after restart")
	restarted := NewView(a, after, saved, func() {}, func() {}, save)
	if restarted.settingsPage.Items[0].Text != "外観" {
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
			s.Language = lang.Code
			w := a.NewWindow("Languages")
			defer w.Close()
			v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
			v.ConfigureUpdates("3.1.0", true, func() {}, func() {})
			v.SetUpdateState(UpdateState{Stage: UpdateAvailable, Version: "3.2.0", CanInstall: true})
			v.Update([]quota.Window{{Remaining: 57, Minutes: 300}, {Remaining: 82, Minutes: 10080}}, i18n.Text(lang.Code, "updated", "14:25"), i18n.Text(lang.Code, "requestOK"))
			w.Resize(fyne.NewSize(420, 600))
			w.Show()
			v.showMain()
			checkButtons(t, v.mainPage)
			v.showSettings()
			for page := range v.settingsPage.Items {
				v.settingsPage.SelectIndex(page)
				img := w.Canvas().Capture()
				if img.Bounds().Empty() {
					t.Fatalf("page %d did not render", page)
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
		})
	}
}
