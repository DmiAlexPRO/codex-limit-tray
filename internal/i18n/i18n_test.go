package i18n

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCatalogs(t *testing.T) {
	if len(catalogs) != len(Languages()) {
		t.Fatal("duplicate language code")
	}
	format := regexp.MustCompile(`%[^% ]*?[dfs]`)
	names := make(map[string]bool)
	for _, language := range Languages() {
		t.Run(language.Code, func(t *testing.T) {
			if language.Name == "" || names[language.Name] {
				t.Fatalf("missing or duplicate language name %q", language.Name)
			}
			names[language.Name] = true
			messages := catalogs[language.Code]
			for key := range messages {
				if _, ok := catalogs[English][key]; !ok {
					t.Errorf("unknown translation key: %s", key)
				}
			}
			for key, english := range catalogs[English] {
				value := messages[key]
				if strings.TrimSpace(value) == "" || !utf8.ValidString(value) {
					t.Errorf("missing or invalid translation: %s", key)
				}
				if strings.ContainsAny(value, "\uFFFD\uFEFF") {
					t.Errorf("unexpected replacement character or BOM: %s", key)
				}
				specifiers := format.FindAllString(english, -1)
				if !reflect.DeepEqual(format.FindAllString(value, -1), specifiers) {
					t.Errorf("format mismatch: %s = %q", key, value)
				}
				args := make([]any, 0, len(specifiers))
				for _, specifier := range specifiers {
					switch specifier[len(specifier)-1] {
					case 'd':
						args = append(args, 17)
					case 'f':
						args = append(args, 17.0)
					case 's':
						args = append(args, "14:25")
					}
				}
				if strings.Contains(Text(language.Code, key, args...), "%!") {
					t.Errorf("format error at runtime: %s = %q", key, value)
				}
			}
		})
	}
}

func TestMatch(t *testing.T) {
	for _, tc := range []struct {
		tags []string
		want string
	}{
		{nil, "en"},
		{[]string{"en-GB", "ru"}, "en"},
		{[]string{"ko-KR", "RU_ru.UTF-8"}, "ko"},
		{[]string{"cs-CZ"}, "cs"},
		{[]string{"zz-ZZ", "RU_ru.UTF-8"}, "ru"},
		{[]string{"pt-BR"}, "pt"},
		{[]string{"ar-EG"}, "arz"},
		{[]string{"ar-SA"}, "ar"},
		{[]string{"zh-Hans-CN"}, "zh"},
		{[]string{"zh-Hant-HK", "ja-JP"}, "ja"},
		{[]string{"pcm-NG"}, "pcm"},
		{[]string{"ur-PK", "ar"}, "ur"},
		{[]string{"be-BY"}, "be"},
		{[]string{"kk-KZ"}, "kk"},
		{[]string{"uk-UA"}, "uk"},
		{[]string{"pl-PL"}, "pl"},
		{[]string{"it-CH"}, "it"},
		{[]string{"tr-TR"}, "tr"},
		{[]string{"ky-KG"}, "ky"},
		{[]string{"hy-AM", "en-US"}, "hy"},
		{[]string{"te-IN"}, "en"},
		{[]string{"zz-ZZ"}, "en"},
	} {
		if got := Match(tc.tags...); got != tc.want {
			t.Errorf("Match(%v) = %q, want %q", tc.tags, got, tc.want)
		}
	}
	if Text("zz", "language") != "Language" || ErrorText("ru", MessageError("login")) != "Войдите в Codex через ChatGPT" {
		t.Fatal("unexpected translation fallback or error")
	}
}
