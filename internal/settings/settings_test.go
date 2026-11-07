package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s := Default()
	s.Theme = "dark"
	s.SoundEnabled = true
	s.AutoStart = true
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	s.Style = "rings"
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil || got != s {
		t.Fatalf("%+v %v", got, err)
	}
	if err = os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = Load(path)
	if err == nil || got != Default() {
		t.Fatal("corruption not recovered")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "broken" {
		t.Fatal("source settings overwritten")
	}
}
func TestValidation(t *testing.T) {
	s := Default()
	s.Threshold = 101
	if s.Validate() == nil {
		t.Fatal("bad threshold")
	}
	for _, v := range []string{"FF", "GGGGGG", "-12345", "1234567"} {
		if _, err := Color(v); err == nil {
			t.Fatalf("bad color accepted: %s", v)
		}
	}
	if c, err := Color("66d9c1"); err != nil || c.G != 217 {
		t.Fatal("valid color rejected")
	}
}

func TestAutoUpdateDefaultsForOldSettingsAndRespectsOptOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	for _, tc := range []struct {
		json string
		want bool
	}{
		{`{"theme":"dark"}`, true},
		{`{"theme":"dark","autoUpdate":false}`, false},
	} {
		if err := os.WriteFile(path, []byte(tc.json), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if err != nil || got.AutoUpdate != tc.want || got.Theme != "dark" {
			t.Fatalf("load %s: %+v, %v", tc.json, got, err)
		}
	}
}
