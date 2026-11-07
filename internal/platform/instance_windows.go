package platform

import (
	"errors"
	"sync/atomic"

	"golang.org/x/sys/windows"
)

// A per-user, per-session event both owns the tray instance and wakes it
// when the user opens another shortcut. No socket or writable lockfile.
func OwnInstance(show func()) (release func(), primary bool, err error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, false, err
	}
	return ownEvent(`Local\CodexLimitTray.Open.`+user.User.Sid.String(), show)
}

func ownEvent(id string, show func()) (func(), bool, error) {
	name, err := windows.UTF16PtrFromString(id)
	if err != nil {
		return nil, false, err
	}
	event, err := windows.CreateEvent(nil, 0, 0, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		err = windows.SetEvent(event)
		windows.CloseHandle(event)
		return nil, false, err
	}
	if err != nil {
		return nil, false, err
	}
	var stopping atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, err := windows.WaitForSingleObject(event, windows.INFINITE)
			if err != nil || stopping.Load() {
				return
			}
			show()
		}
	}()
	return func() {
		stopping.Store(true)
		windows.SetEvent(event)
		<-done
		windows.CloseHandle(event)
	}, true, nil
}
