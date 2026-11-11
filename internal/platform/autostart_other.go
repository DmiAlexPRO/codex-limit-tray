//go:build !windows && !darwin

package platform

import "errors"

func AutoStartSupported() bool { return false }

func SetAutoStart(bool) error {
	return errors.New("start at login is supported on Windows and macOS only")
}
