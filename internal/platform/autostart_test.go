package platform

import "testing"

func TestAutoStartArguments(t *testing.T) {
	arguments := autoStartArguments(`C:\Apps\Codex Limit Tray.exe`)
	if len(arguments) != 2 || arguments[0] != `C:\Apps\Codex Limit Tray.exe` || arguments[1] != "--minimized" {
		t.Fatalf("unexpected arguments: %#v", arguments)
	}
}
