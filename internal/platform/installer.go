package platform

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

func Installed() bool {
	path, err := os.Executable()
	if err != nil || runtime.GOOS != "windows" {
		return false
	}
	marker, err := os.Lstat(filepath.Join(filepath.Dir(path), "installed.flag"))
	return err == nil && marker.Mode().IsRegular()
}

// Only call with a local path returned by the signed update verifier.
func LaunchInstaller(path string) error {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" {
		return errors.New("in-app installation is only available on Windows x64")
	}
	cmd := exec.Command(path, "/SP-", "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/UPDATE",
		"/WAITPID="+strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return errors.New("Windows could not launch the verified installer")
	}
	// The installer waits for this PID and then owns the restart.
	_ = cmd.Process.Release()
	return nil
}
