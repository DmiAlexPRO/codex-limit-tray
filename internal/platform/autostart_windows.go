package platform

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const windowsRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func AutoStartSupported() bool { return true }

func SetAutoStart(enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, windowsRunKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("could not open start-at-login settings: %w", err)
	}
	defer key.Close()
	if !enabled {
		if err = key.DeleteValue(autoStartLabel); err != nil && !errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
			return fmt.Errorf("could not disable start at login: %w", err)
		}
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate the app: %w", err)
	}
	arguments := autoStartArguments(executable)
	command := syscall.EscapeArg(arguments[0]) + " " + arguments[1]
	if err = key.SetStringValue(autoStartLabel, command); err != nil {
		return fmt.Errorf("could not enable start at login: %w", err)
	}
	return nil
}
