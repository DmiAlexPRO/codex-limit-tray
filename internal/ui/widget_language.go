package ui

import (
	"sync/atomic"

	"codexlimittray/internal/i18n"
	"fyne.io/fyne/v2/lang"
)

var widgetLanguage atomic.Value

// Fyne uses lang.L for its built-in menus and dialogs. Resolve the strings
// this app uses against the app language, rather than the OS language.
func init() {
	keys := map[string]string{
		"Cancel": "earnedResetCancel", "OK": "ok", "Error": "error",
		"Advanced": "colorAdvanced", "Confirm": "confirm",
		"Cut": "cut", "Copy": "copy", "Paste": "paste", "Select all": "selectAll",
		"Undo": "undo", "Redo": "redo",
	}
	fallback := lang.L
	lang.L = func(label string, data ...any) string {
		key := keys[label]
		if key == "" {
			return fallback(label, data...)
		}
		language, _ := widgetLanguage.Load().(string)
		return i18n.Text(language, key)
	}
}

func setWidgetLanguage(language string) { widgetLanguage.Store(language) }
