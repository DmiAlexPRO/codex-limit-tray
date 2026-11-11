package ui

import (
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenderThemesAndScales(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	for _, mode := range []string{"light", "dark"} {
		for _, scale := range []float32{1, 1.5, 2} {
			s := settings.Default()
			s.Theme = mode
			w := a.NewWindow("Test")
			v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
			v.Update([]quota.Window{{Remaining: 82, Minutes: 300, Reset: time.Date(2026, 11, 5, 13, 18, 0, 0, time.UTC)}, {Remaining: 57, Minutes: 10080, Reset: time.Date(2026, 11, 5, 13, 18, 0, 0, time.UTC)}}, "Обновлено 18:41:49 · запрос каждые 30 с", "Fixture")
			c := w.Canvas().(software.WindowlessCanvas)
			c.SetScale(scale)
			w.Resize(fyne.NewSize(420, 600))
			w.Show()
			img := w.Canvas().Capture()
			controls := findWindowControls(w.Content())
			if len(controls) != 2 {
				t.Fatalf("expected two window controls, got %d", len(controls))
			}
			for _, control := range controls {
				pos := a.Driver().AbsolutePositionForObject(control)
				if pos.Y < 12 || pos.X+control.Size().Width > w.Canvas().Size().Width-12 {
					t.Fatalf("control too close to rounded edge: pos=%v size=%v canvas=%v", pos, control.Size(), w.Canvas().Size())
				}
			}
			if img.Bounds().Dx() < 400 {
				t.Fatal("capture too small")
			}
			checkButtons(t, v.dashboard)
			if os.Getenv("CODEX_LIMIT_PREVIEW") == "1" {
				path := filepath.Join("..", "..", "dist", "fyne-preview")
				if err := os.MkdirAll(path, 0755); err != nil {
					t.Fatal(err)
				}
				f, err := os.Create(filepath.Join(path, fmt.Sprintf("dashboard-%s-%g.png", mode, scale)))
				if err != nil {
					t.Fatal(err)
				}
				if err = png.Encode(f, img); err != nil {
					t.Fatal(err)
				}
				f.Close()
			}
			v.Update(nil, "Нет данных", "Fixture")
			if len(v.windows) != 0 {
				t.Fatal("stale values")
			}
			v.Update([]quota.Window{{Remaining: 100, Minutes: 300}, {Remaining: 0, Minutes: 10080}}, "Fixture", "Fixture")
			checkButtons(t, v.dashboard)
			w.Close()
		}
	}
}

func findWindowControls(obj fyne.CanvasObject) []*windowControl {
	if c, ok := obj.(*windowControl); ok {
		return []*windowControl{c}
	}
	var found []*windowControl
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			found = append(found, findWindowControls(child)...)
		}
	}
	return found
}
func checkButtons(t *testing.T, obj fyne.CanvasObject) {
	t.Helper()
	if b, ok := obj.(*widget.Button); ok && b.Visible() {
		if b.Size().Width < b.MinSize().Width || b.Size().Height < b.MinSize().Height {
			t.Errorf("clipped button %q: %v < %v", b.Text, b.Size(), b.MinSize())
		}
	}
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			checkButtons(t, child)
		}
	}
}
func TestIndicators(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	s := settings.Default()
	s.ThresholdEnabled = true
	if ValueColor(s, quota.Window{Remaining: 19}, 0) != hex(s.LowColor) || ValueColor(s, quota.Window{Remaining: 20}, 0) != hex(s.Accent) {
		t.Fatal("threshold boundary")
	}
	weekly := quota.Window{Remaining: 57, Minutes: 10080}
	if ValueColor(s, weekly, 0) != ValueColor(s, weekly, 1) || ValueColor(s, weekly, 0) == hex(s.Accent) {
		t.Fatal("weekly color must not depend on presence of five-hour quota")
	}
	for _, style := range []string{"numbers", "bars", "segments", "rings"} {
		s.Style = style
		for _, windows := range [][]quota.Window{nil, {{Remaining: 0}}, {{Remaining: 100}}, {{Remaining: 57}, {Remaining: 82}}} {
			img := Indicator(windows, s)
			if img.Bounds().Dx() != 128 {
				t.Fatal("invalid icon size")
			}
		}
	}
}

func TestSettingsPagesAndThemeSwitch(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	s := settings.Default()
	w := a.NewWindow("Test")
	v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
	v.ConfigureUpdates("1.0.0", true, func() {}, func() {})
	v.SetUpdateState(UpdateState{Stage: UpdateAvailable, Version: "1.1.0", Notes: "Улучшена читаемость индикатора.", CanInstall: true})
	v.showSettings()
	tabs := findTabs(w.Content())
	if tabs == nil {
		t.Fatal("tabs missing")
	}
	w.Show()
	for _, mode := range []string{"light", "dark"} {
		s.Theme = mode
		a.Settings().SetTheme(appTheme{s})
		w.Content().Refresh()
		for _, page := range []int{0, 1, 2} {
			tabs.SelectIndex(page)
			for _, scale := range []float32{1, 1.5, 2} {
				w.Canvas().(software.WindowlessCanvas).SetScale(scale)
				w.Resize(fyne.NewSize(420, 600))
				img := w.Canvas().Capture()
				if img.Bounds().Empty() {
					t.Fatal("empty page")
				}
				if os.Getenv("CODEX_LIMIT_PREVIEW") != "1" {
					continue
				}
				path := filepath.Join("..", "..", "dist", "fyne-preview")
				if err := os.MkdirAll(path, 0755); err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(filepath.Join(path, fmt.Sprintf("page-%d-%s-%g.png", page, mode, scale)))
				if err != nil {
					t.Fatal(err)
				}
				if err := png.Encode(file, img); err != nil {
					t.Fatal(err)
				}
				file.Close()
			}
		}
	}
}

func findTabs(obj fyne.CanvasObject) *container.AppTabs {
	if tabs, ok := obj.(*container.AppTabs); ok {
		return tabs
	}
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if tabs := findTabs(child); tabs != nil {
				return tabs
			}
		}
	}
	return nil
}

func findForm(obj fyne.CanvasObject) *widget.Form {
	if form, ok := obj.(*widget.Form); ok {
		return form
	}
	if scroll, ok := obj.(*container.Scroll); ok {
		return findForm(scroll.Content)
	}
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if form := findForm(child); form != nil {
				return form
			}
		}
	}
	return nil
}

func TestDropdownDoesNotResizeCanvas(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	v := NewView(a, w, settings.Default(), func() {}, func() {}, func(settings.Settings) error { return nil })
	content := v.preferences()
	w.SetContent(content)
	w.Resize(fyne.NewSize(420, 600))
	w.Show()
	form := findForm(content)
	if form == nil {
		t.Fatal("settings form not found")
	}
	for _, index := range []int{0, 1, 2} {
		selectBox, ok := form.Items[index].Widget.(*widget.Select)
		if !ok {
			t.Fatal("select not found")
		}
		before := w.Canvas().Size()
		test.Tap(selectBox)
		if w.Canvas().Overlays().Top() == nil {
			t.Fatal("dropdown did not open")
		}
		if after := w.Canvas().Size(); after != before {
			t.Fatalf("dropdown resized canvas: %v -> %v", before, after)
		}
		w.Canvas().Overlays().Top().Hide()
	}
}

func TestCompactIndicator(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("widget")
	hidden := false
	i := newIndicatorWidget(w, WidgetIndicator([]quota.Window{{Remaining: 57, Minutes: 10080}}, settings.Default()), func() { hidden = true })
	w.SetPadded(false)
	w.SetContent(i)
	w.Resize(fyne.NewSize(184, 88))
	w.Show()
	if i.MinSize() != fyne.NewSize(184, 88) {
		t.Fatal(i.MinSize())
	}
	i.MouseIn(nil)
	if !i.hover {
		t.Fatal("hover state missing")
	}
	i.Tapped(&fyne.PointEvent{Position: fyne.NewPos(19, 17)})
	if !hidden {
		t.Fatal("hover close did not hide widget")
	}
	hidden = false
	i.MouseOut()
	if i.hover {
		t.Fatal("hover not cleared")
	}
	i.Tapped(&fyne.PointEvent{Position: fyne.NewPos(19, 17)})
	if hidden {
		t.Fatal("invisible close remained clickable")
	}
	i.TappedSecondary(&fyne.PointEvent{})
	if !hidden {
		t.Fatal("secondary click must hide the widget")
	}
	if os.Getenv("CODEX_LIMIT_PREVIEW") == "1" {
		f, err := os.Create(filepath.Join("..", "..", "dist", "fyne-preview", "widget.png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, w.Canvas().Capture()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTooltipAndWidgetStyles(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	windows := []quota.Window{{Remaining: 57, Minutes: 300}, {Remaining: 12, Minutes: 10080}}
	if QuotaTooltip(windows, "ru") != "Codex Limits\n5 часов: 57%\nНеделя: 12%" {
		t.Fatal(QuotaTooltip(windows, "ru"))
	}
	if QuotaTooltip(nil, "ru") != "Codex Limits — Данные не получены" {
		t.Fatal("unknown tooltip")
	}
	for _, style := range []string{"numbers", "bars", "segments", "rings"} {
		s := settings.Default()
		s.Style = style
		img := WidgetIndicator(windows, s)
		if img.Bounds().Dx() != 368 || img.Bounds().Dy() != 176 {
			t.Fatal(img.Bounds())
		}
	}
}
