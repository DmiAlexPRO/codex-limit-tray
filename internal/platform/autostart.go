package platform

const autoStartLabel = "CodexLimitTray"

func autoStartArguments(executable string) []string {
	return []string{executable, "--minimized"}
}
