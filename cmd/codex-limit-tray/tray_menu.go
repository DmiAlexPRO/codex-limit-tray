package main

import (
	"codexlimittray/internal/i18n"

	"fyne.io/fyne/v2"
)

func trayMenuUpdater(setMenu func(*fyne.Menu), show, refresh, showWidget, openSettings, exit func()) func(string) {
	return func(language string) {
		setMenu(newTrayMenu(language, show, refresh, showWidget, openSettings, exit))
	}
}

func newTrayMenu(language string, show, refresh, showWidget, openSettings, exit func()) *fyne.Menu {
	item := func(key string, action func()) *fyne.MenuItem {
		return fyne.NewMenuItem(i18n.Text(language, key), action)
	}
	items := []*fyne.MenuItem{
		item("show", show),
		item("refresh", refresh),
		item("showWidget", showWidget),
		item("settings", openSettings),
	}
	quit := item("exit", exit)
	quit.IsQuit = true // Fyne adds its own Quit unless the final item is marked.
	return fyne.NewMenu("Codex Limits", append(items, fyne.NewMenuItemSeparator(), quit)...)
}
