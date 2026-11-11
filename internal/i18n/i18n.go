// Package i18n contains the bundled app translations and startup locale matching.
package i18n

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jeandeaual/go-locale"
)

const English = "en"

type Language struct {
	Code string
	Name string
}

// Languages returns the display order of the language picker.
func Languages() []Language {
	return []Language{
		{"en", "English"}, {"zh", "简体中文"}, {"hi", "हिन्दी"},
		{"es", "Español"}, {"ar", "العربية الفصحى"}, {"fr", "Français"},
		{"bn", "বাংলা"}, {"pt", "Português"}, {"id", "Bahasa Indonesia"},
		{"ur", "اردو"}, {"ru", "Русский"}, {"de", "Deutsch"},
		{"ja", "日本語"}, {"pcm", "Naijá"}, {"arz", "العربية المصرية"},
		{"mr", "मराठी"}, {"vi", "Tiếng Việt"},
		{"sw", "Kiswahili"}, {"ha", "Hausa"},
		{"be", "Беларуская"}, {"kk", "Қазақша"}, {"uk", "Українська"},
		{"pl", "Polski"}, {"it", "Italiano"}, {"tr", "Türkçe"},
		{"ky", "Кыргызча"}, {"hy", "Հայերեն"},
	}
}

//go:embed locales/*.json
var files embed.FS

var catalogs = loadCatalogs()

func loadCatalogs() map[string]map[string]string {
	catalogs := make(map[string]map[string]string)
	for _, lang := range Languages() {
		data, err := files.ReadFile("locales/" + lang.Code + ".json")
		if err != nil {
			panic(err)
		}
		var messages map[string]string
		if err := json.Unmarshal(data, &messages); err != nil {
			panic(err)
		}
		catalogs[lang.Code] = messages
	}
	return catalogs
}

func Supported(code string) bool { return catalogs[code] != nil }

func Match(locales ...string) string {
	for _, tag := range locales {
		tag = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(tag)), "_", "-")
		tag, _, _ = strings.Cut(tag, ".")
		tag, _, _ = strings.Cut(tag, "@")
		base, _, _ := strings.Cut(tag, "-")
		switch {
		case strings.HasPrefix(tag, "zh-hant") || tag == "zh-tw" || tag == "zh-hk" || tag == "zh-mo":
			continue // Only Simplified Chinese is bundled.
		case tag == "ar-eg":
			base = "arz"
		case base == "cmn":
			base = "zh"
		case base == "arb":
			base = "ar"
		}
		if Supported(base) {
			return base
		}
	}
	return English
}

func Detect() string {
	if preferred := preferredLocales(); len(preferred) > 0 {
		return Match(preferred...)
	}
	locales, _ := locale.GetLocales()
	return Match(locales...)
}

func Text(code, key string, args ...any) string {
	value := catalogs[code][key]
	if value == "" {
		value = catalogs[English][key]
	}
	if value == "" {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(value, args...)
	}
	return value
}

// MessageError preserves a stable message key across background and UI threads.
type MessageError string

func (e MessageError) Error() string { return Text(English, string(e)) }

func ErrorText(code string, err error) string {
	var localized MessageError
	if errors.As(err, &localized) {
		return Text(code, string(localized))
	}
	return err.Error()
}
