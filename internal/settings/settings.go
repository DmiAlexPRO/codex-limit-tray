package settings

import (
	"codexlimittray/internal/i18n"
	"encoding/json"
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
)

type Settings struct {
	Language         string `json:"language"`
	Theme            string `json:"theme"`
	Style            string `json:"style"`
	Accent           string `json:"accent"`
	LowColor         string `json:"lowColor"`
	Threshold        int    `json:"threshold"`
	ThresholdEnabled bool   `json:"thresholdEnabled"`
	SoundEnabled     bool   `json:"soundEnabled"`
	AutoStart        bool   `json:"autoStart"`
	AutoUpdate       bool   `json:"autoUpdate"`
	CodexPath        string `json:"codexPath"`
}

func Default() Settings {
	return Settings{Language: i18n.English, Theme: "system", Style: "numbers", Accent: "66D9C1", LowColor: "FA7185", Threshold: 20, AutoUpdate: true}
}
func (s Settings) Validate() error {
	if !i18n.Supported(s.Language) {
		return errors.New("unsupported language")
	}
	if s.Theme != "system" && s.Theme != "light" && s.Theme != "dark" {
		return errors.New("unknown theme")
	}
	if s.Style != "numbers" && s.Style != "bars" && s.Style != "segments" && s.Style != "rings" {
		return errors.New("unknown icon style")
	}
	if _, err := Color(s.Accent); err != nil {
		return err
	}
	if _, err := Color(s.LowColor); err != nil {
		return err
	}
	if s.Threshold < 1 || s.Threshold > 100 {
		return errors.New("threshold must be between 1 and 100")
	}
	if s.CodexPath != "" && !filepath.IsAbs(s.CodexPath) {
		return errors.New("specify an absolute Codex path")
	}
	return nil
}
func Color(value string) (color.NRGBA, error) {
	if len(value) != 6 {
		return color.NRGBA{}, errors.New("color must be a 6-character HEX value, e.g. 66D9C1")
	}
	n, err := strconv.ParseUint(value, 16, 24)
	if err != nil {
		return color.NRGBA{}, errors.New("invalid HEX color")
	}
	return color.NRGBA{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n), A: 255}, nil
}
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "CodexLimitTray", "fyne-preview.json"), nil
}

// InitLanguage seeds only the language on a fresh install, including silent
// installs. An upgrade, reinstall or portable user may already have settings.
func InitLanguage(path, language string) error {
	if !i18n.Supported(language) {
		return errors.New("unsupported language")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(map[string]string{"language": language})
	if err = errors.Join(err, f.Close()); err != nil {
		_ = os.Remove(path)
	}
	return err
}

func Load(path, defaultLanguage string) (Settings, error) {
	defaults := Default()
	if i18n.Supported(defaultLanguage) {
		defaults.Language = defaultLanguage
	}
	s := defaults
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if info.Size() > 32768 {
		return s, errors.New("settings file is too large")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(data, &s); err != nil {
		return defaults, errors.New("invalid settings file")
	}
	if !i18n.Supported(s.Language) {
		// A language from a newer version must not reset unrelated settings.
		s.Language = defaults.Language
	}
	if err = s.Validate(); err != nil {
		return defaults, err
	}
	return s, nil
}
func (s Settings) Save(path string) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), "settings-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
