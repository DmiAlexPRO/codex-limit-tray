package settings

import (
	"encoding/json"
	"errors"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
)

type Settings struct {
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
	return Settings{Theme: "system", Style: "numbers", Accent: "66D9C1", LowColor: "FA7185", Threshold: 20, AutoUpdate: true}
}
func (s Settings) Validate() error {
	if s.Theme != "system" && s.Theme != "light" && s.Theme != "dark" {
		return errors.New("неизвестная тема")
	}
	if s.Style != "numbers" && s.Style != "bars" && s.Style != "segments" && s.Style != "rings" {
		return errors.New("неизвестный стиль")
	}
	if _, err := Color(s.Accent); err != nil {
		return err
	}
	if _, err := Color(s.LowColor); err != nil {
		return err
	}
	if s.Threshold < 1 || s.Threshold > 100 {
		return errors.New("порог должен быть от 1 до 100")
	}
	if s.CodexPath != "" && !filepath.IsAbs(s.CodexPath) {
		return errors.New("укажите абсолютный путь Codex")
	}
	return nil
}
func Color(value string) (color.NRGBA, error) {
	if len(value) != 6 {
		return color.NRGBA{}, errors.New("цвет: 6 символов HEX, например 66D9C1")
	}
	n, err := strconv.ParseUint(value, 16, 24)
	if err != nil {
		return color.NRGBA{}, errors.New("некорректный HEX-цвет")
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
func Load(path string) (Settings, error) {
	s := Default()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if info.Size() > 32768 {
		return s, errors.New("файл настроек слишком большой")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(data, &s); err != nil {
		return Default(), errors.New("настройки повреждены; используются стандартные")
	}
	if err = s.Validate(); err != nil {
		return Default(), err
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
