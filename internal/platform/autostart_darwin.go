package platform

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

const launchAgentLabel = "io.codexlimittray.preview"

func AutoStartSupported() bool { return true }

func launchAgentPlist(executable string) (string, error) {
	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, []byte(executable)); err != nil {
		return "", err
	}
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>` + launchAgentLabel + `</string>
<key>ProgramArguments</key><array>
<string>` + escaped.String() + `</string>
<string>--minimized</string>
</array>
<key>RunAtLoad</key><true/>
<key>ProcessType</key><string>Background</string>
</dict></plist>
`, nil
}

func SetAutoStart(enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %w", err)
	}
	path := filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist")
	if !enabled {
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not disable start at login: %w", err)
		}
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate the app: %w", err)
	}
	content, err := launchAgentPlist(executable)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("could not create LaunchAgents: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "codex-limit-tray-*.plist")
	if err != nil {
		return fmt.Errorf("could not save start-at-login settings: %w", err)
	}
	defer os.Remove(temporary.Name())
	if err = temporary.Chmod(0644); err == nil {
		_, err = temporary.WriteString(content)
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(temporary.Name(), path)
	}
	if err != nil {
		return fmt.Errorf("could not enable start at login: %w", err)
	}
	return nil
}
