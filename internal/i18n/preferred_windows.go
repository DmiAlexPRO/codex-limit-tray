package i18n

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows display languages take priority over the locale used for formatting.
func preferredLocales() []string {
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetUserPreferredUILanguages")
	var count, size uint32
	const nameFormat = 8 // MUI_LANGUAGE_NAME
	ok, _, _ := proc.Call(nameFormat, uintptr(unsafe.Pointer(&count)), 0, uintptr(unsafe.Pointer(&size)))
	if ok == 0 || size < 2 || size > 32768 {
		return nil
	}
	buf := make([]uint16, size)
	ok, _, _ = proc.Call(nameFormat, uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ok == 0 {
		return nil
	}
	var locales []string
	for len(buf) > 0 && buf[0] != 0 {
		end := 0
		for end < len(buf) && buf[end] != 0 {
			end++
		}
		locales = append(locales, windows.UTF16ToString(buf[:end]))
		if end == len(buf) {
			break
		}
		buf = buf[end+1:]
	}
	return locales
}
