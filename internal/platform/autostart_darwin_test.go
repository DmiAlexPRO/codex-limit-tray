package platform

import (
	"strings"
	"testing"
)

func TestLaunchAgentPlist(t *testing.T) {
	content, err := launchAgentPlist("/Applications/Codex & Limits.app/Contents/MacOS/codex-limit-tray")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"<key>RunAtLoad</key><true/>",
		"<string>--minimized</string>",
		"<string>/Applications/Codex &amp; Limits.app/Contents/MacOS/codex-limit-tray</string>",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("plist does not contain %q:\n%s", expected, content)
		}
	}
}
