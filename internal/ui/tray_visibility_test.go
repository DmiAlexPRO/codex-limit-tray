package ui

import (
	"errors"
	"net/url"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

type settingsLinkApp struct {
	fyne.App
	opened string
	err    error
}

func (a *settingsLinkApp) OpenURL(link *url.URL) error {
	a.opened = link.String()
	return a.err
}

func TestTrayVisibilityHelp(t *testing.T) {
	a := &settingsLinkApp{App: test.NewApp()}
	defer a.Quit()
	w := a.NewWindow("Tray visibility")
	v := &View{app: a, window: w, language: "ru"}
	windows := v.trayVisibilityHelp("windows")
	w.SetContent(container.NewVBox(windows...))
	w.Resize(fyne.NewSize(400, 240))
	w.Show()
	if w.Canvas().Capture().Bounds().Empty() {
		t.Fatal("tray help did not render")
	}
	hint, ok := windows[1].(*widget.Label)
	if !ok || hint.Text != v.tr("taskbarIconHelp") {
		t.Fatal("missing Windows tray instructions")
	}
	button, ok := windows[2].(*widget.Button)
	if !ok || button.Text != "Открыть параметры панели задач" {
		t.Fatal("missing localized Windows settings button")
	}
	checkButtons(t, w.Content())
	test.Tap(button)
	if a.opened != "ms-settings:taskbar" || w.Canvas().Overlays().Top() != nil {
		t.Fatal("button did not open the Windows taskbar settings URI")
	}
	a.err = errors.New("settings unavailable")
	test.Tap(button)
	if w.Canvas().Overlays().Top() == nil {
		t.Fatal("failed settings launch did not show a fallback")
	}

	mac := v.trayVisibilityHelp("darwin")
	if len(mac) != 2 {
		t.Fatal("expected only a separator and menu bar instructions on macOS")
	}
	hint, ok = mac[1].(*widget.Label)
	if !ok || hint.Text != v.tr("menuBarIconHelp") {
		t.Fatal("missing macOS menu bar instructions")
	}
	if len(v.trayVisibilityHelp("linux")) != 0 {
		t.Fatal("unexpected platform instructions on Linux")
	}
}
