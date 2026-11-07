//go:build !windows

package update

import "os"

func reparse(info os.FileInfo) bool { return info.Mode()&os.ModeSymlink != 0 }
