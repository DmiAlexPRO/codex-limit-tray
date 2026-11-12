package main

import (
	"slices"
	"testing"

	"codexlimittray/internal/i18n"

	"fyne.io/fyne/v2"
)

func TestTrayMenu(t *testing.T) {
	for _, lang := range i18n.Languages() {
		t.Run(lang.Code, func(t *testing.T) {
			var called []string
			action := func(name string) func() {
				return func() { called = append(called, name) }
			}
			menu := newTrayMenu(lang.Code, action("show"), action("refresh"), action("widget"), action("settings"), action("quit"))
			if len(menu.Items) != 6 || !menu.Items[4].IsSeparator {
				t.Fatal("expected four actions, a separator, and quit")
			}
			seen := make(map[string]bool)
			for index, item := range menu.Items {
				if item.IsQuit != (index == len(menu.Items)-1) {
					t.Fatal("only the final item must be marked as quit")
				}
				if item.IsSeparator {
					continue
				}
				if item.Label == "" || seen[item.Label] || item.Action == nil {
					t.Fatalf("invalid or duplicate tray action %q", item.Label)
				}
				seen[item.Label] = true
				item.Action()
			}
			if !slices.Equal(called, []string{"show", "refresh", "widget", "settings", "quit"}) {
				t.Fatalf("wrong tray action mapping: %v", called)
			}
		})
	}
}

func TestTrayMenuUpdatesWithoutRestart(t *testing.T) {
	var menu *fyne.Menu
	quitCalls := 0
	action := func() {}
	updateMenu := trayMenuUpdater(func(next *fyne.Menu) { menu = next }, action, action, action, action, func() { quitCalls++ })
	updateMenu("en")
	if menu == nil || len(menu.Items) != 6 || menu.Items[3].Label != "Settings" {
		t.Fatal("initial menu not installed")
	}
	updateMenu("ru")
	if menu == nil || len(menu.Items) != 6 || menu.Items[3].Label != "Настройки" || !menu.Items[5].IsQuit {
		t.Fatal("localized tray menu not installed")
	}
	menu.Items[5].Action()
	if quitCalls != 1 {
		t.Fatal("quit action lost on language change")
	}
}

func TestRussianTrayMenuLabels(t *testing.T) {
	action := func() {}
	menu := newTrayMenu("ru", action, action, action, action, action)
	labels := make([]string, 0, len(menu.Items))
	for _, item := range menu.Items {
		labels = append(labels, item.Label)
	}
	want := []string{"Открыть окно", "Обновить лимиты", "Показать виджет", "Настройки", "", "Выйти из Codex Limit Tray"}
	if !slices.Equal(labels, want) {
		t.Errorf("labels = %q, want %q", labels, want)
	}
}
