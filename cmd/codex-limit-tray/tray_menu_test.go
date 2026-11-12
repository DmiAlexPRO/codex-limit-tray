package main

import (
	"slices"
	"testing"

	"codexlimittray/internal/i18n"
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
