//go:build !windows && !darwin

package platform

import "errors"

func AutoStartSupported() bool { return false }

func SetAutoStart(bool) error {
	return errors.New("автозапуск поддерживается только на Windows и macOS")
}
