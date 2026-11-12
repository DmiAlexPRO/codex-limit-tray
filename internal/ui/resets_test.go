package ui

import (
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestResetPanelConfirmation(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	s := settings.Default()
	s.Theme, s.Language = "dark", "ru"
	w := a.NewWindow("Test")
	v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
	defer w.Close()
	v.Update([]quota.Window{{Remaining: 18, Minutes: 10080, Reset: time.Date(2026, 11, 12, 17, 2, 0, 0, time.UTC)}}, "Данные для проверки интерфейса", "")
	w.Resize(fyne.NewSize(420, 600))
	w.Show()
	requests := 0
	v.ConfigureReset(func() { requests++ })
	p := v.resets
	if p.button.Text != "Сбросить…" {
		t.Fatalf("expected the short reset label, got %q", p.button.Text)
	}
	if p.body.Visible() {
		t.Fatal("unknown credits must not appear as zero")
	}
	v.SetResetCredits(&quota.ResetCredits{}, "")
	if !p.body.Visible() || !p.button.Disabled() {
		t.Fatal("zero credits should display without allowing redemption")
	}
	credits := &quota.ResetCredits{Available: 3, ExpiresAt: []time.Time{time.Date(2026, 12, 6, 12, 0, 0, 0, time.UTC), {}}}
	v.SetResetCredits(credits, "")
	if len(p.body.Objects) != 3 || p.button.Disabled() {
		t.Fatal("expected an enabled button and compact list")
	}
	rows, ok := p.body.Objects[2].(*widget.Label)
	if !ok || strings.Count(rows.Text, "\n") != 1 || !strings.Contains(rows.Text, "До "+credits.ExpiresAt[0].Local().Format("02.01.2006")) {
		t.Fatal("expected two detail rows, one with expiry")
	}
	saveResetPreview(t, w, "resets.png")
	tapDialog := func(label string) {
		t.Helper()
		overlay, ok := w.Canvas().Overlays().Top().(fyne.Widget)
		if !ok {
			t.Fatal("confirmation did not open")
		}
		button := resetDialogButton(overlay, label)
		if button == nil {
			t.Fatalf("confirmation lacks %q", label)
		}
		test.Tap(button)
	}
	test.Tap(p.button)
	if requests != 0 {
		t.Fatal("credit requested before confirmation")
	}
	saveResetPreview(t, w, "resets-confirm.png")
	tapDialog("Отмена")
	if requests != 0 {
		t.Fatal("cancel requested a reset")
	}
	test.Tap(p.button)
	tapDialog("Использовать")
	v.SetResetCredits(credits, "")
	test.Tap(p.button)
	if requests != 1 || !p.button.Disabled() {
		t.Fatal("expected exactly one pending redemption")
	}
	v.SetResetCredits(&quota.ResetCredits{Available: 2}, "earnedResetDone")
	if p.pending || p.button.Disabled() || p.message != "earnedResetDone" {
		t.Fatal("completed redemption did not update the controls")
	}
}

func TestResetPanelScrollsAfterFiveRows(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	defer w.Close()
	s := settings.Default()
	s.Theme, s.Language = "dark", "ru"
	v := NewView(a, w, s, func() {}, func() {}, func(settings.Settings) error { return nil })
	v.ConfigureReset(func() {})
	w.Resize(fyne.NewSize(420, 600))
	w.Show()
	windows := []quota.Window{{Remaining: 18, Minutes: 10080}}
	v.Update(windows, "Данные для проверки интерфейса", "")
	p := v.resets
	setRows := func(count int) {
		v.SetResetCredits(&quota.ResetCredits{Available: count, ExpiresAt: make([]time.Time, count)}, "")
	}
	setRows(5)
	if p.body.Objects[2] != p.rows {
		t.Fatal("five rows should fit without scrolling")
	}
	listHeight := p.rows.Size().Height
	panelHeight, windowSize := p.body.Size().Height, w.Canvas().Size()
	for _, count := range []int{6, 50} {
		setRows(count)
		if p.body.Objects[2] != p.scrollBody || p.scrollBody.Size().Height != listHeight {
			t.Fatalf("%d rows should scroll in a five-row viewport; got height %v, want %v", count, p.scrollBody.Size().Height, listHeight)
		}
		if p.body.Size().Height != panelHeight || w.Canvas().Size() != windowSize {
			t.Fatalf("%d rows expanded the reset panel or the window", count)
		}
	}
	pos := a.Driver().AbsolutePositionForObject(p.scroll).Add(fyne.NewPos(12, 12))
	test.Scroll(w.Canvas(), pos, 0, -10000)
	if want := p.rows.Size().Height - p.scroll.Size().Height; p.scroll.Offset.Y != want || want <= 0 {
		t.Fatalf("scroll did not reach the last row: offset %v, want %v", p.scroll.Offset.Y, want)
	}
	setRows(50)
	v.Update(windows, "Данные для проверки интерфейса", "")
	if p.scroll.Offset.Y <= 0 {
		t.Fatal("refresh reset the scroll position")
	}
	setRows(4)
	if p.body.Objects[2] != p.rows || p.scroll.Offset.Y != 0 {
		t.Fatal("the short list should be visible from the first row without scrolling")
	}
	setRows(8)
	saveResetPreview(t, w, "resets-scroll.png")
}

func resetDialogButton(obj fyne.CanvasObject, label string) *widget.Button {
	if button, ok := obj.(*widget.Button); ok && button.Text == label {
		return button
	}
	var children []fyne.CanvasObject
	switch node := obj.(type) {
	case *fyne.Container:
		children = node.Objects
	case *widget.PopUp:
		children = []fyne.CanvasObject{node.Content}
	case fyne.Widget:
		children = test.WidgetRenderer(node).Objects()
	}
	for _, child := range children {
		if button := resetDialogButton(child, label); button != nil {
			return button
		}
	}
	return nil
}

func saveResetPreview(t *testing.T, w fyne.Window, name string) {
	t.Helper()
	if os.Getenv("CODEX_LIMIT_PREVIEW") != "1" {
		return
	}
	dir := filepath.Join("..", "..", "dist", "fyne-preview")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
