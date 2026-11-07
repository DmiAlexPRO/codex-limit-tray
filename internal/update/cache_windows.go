package update

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func reparse(info os.FileInfo) bool {
	st, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return !ok || st.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
