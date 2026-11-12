package i18n

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCatalogs(t *testing.T) {
	if len(catalogs) != 27 || len(Languages()) != 27 {
		t.Fatalf("want 27 unique languages, got %d", len(catalogs))
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
				value, present := messages[key]
				// Earned reset messages fall back to English outside the initial locales.
				if !present && language.Code != English && language.Code != "ru" && strings.HasPrefix(key, "earnedReset") {
					continue
				}
				if strings.TrimSpace(value) == "" || !utf8.ValidString(value) {
					t.Errorf("missing or invalid translation: %s", key)
				}
				if !reflect.DeepEqual(format.FindAllString(value, -1), format.FindAllString(english, -1)) {
					t.Errorf("format mismatch: %s = %q", key, value)
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
		{[]string{"ko-KR", "RU_ru.UTF-8"}, "ru"},
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
