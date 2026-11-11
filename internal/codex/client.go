package codex

import (
	"bufio"
	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/update"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var ErrLogin = i18n.MessageError("login")
var ErrVersion = i18n.MessageError("version")

type response struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
	Params struct {
		AuthMode *string `json:"authMode"`
	} `json:"params"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func Resolve(override string) (string, error) {
	if override != "" {
		if !filepath.IsAbs(override) {
			return "", i18n.MessageError("codexPath")
		}
		if info, err := os.Stat(override); err == nil && !info.IsDir() {
			return override, nil
		}
		return "", i18n.MessageError("codexMissing")
	}
	if runtime.GOOS == "windows" {
		root := filepath.Join(os.Getenv("LOCALAPPDATA"), "OpenAI", "Codex", "bin")
		var newest string
		var modified time.Time
		_ = filepath.WalkDir(root, func(path string, item os.DirEntry, err error) error {
			if err != nil || item.IsDir() || !strings.EqualFold(item.Name(), "codex.exe") {
				return nil
			}
			if info, statErr := item.Info(); statErr == nil && info.ModTime().After(modified) {
				newest, modified = path, info.ModTime()
			}
			return nil
		})
		if newest != "" {
			return newest, nil
		}
		for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
			candidate := filepath.Join(dir, "node_modules", "@openai", "codex", "node_modules", "@openai", "codex-win32-x64", "vendor", "x86_64-pc-windows-msvc", "codex", "codex.exe")
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	if path, err := exec.LookPath("codex"); err == nil {
		return path, nil
	}
	if runtime.GOOS == "darwin" {
		for _, path := range []string{"/opt/homebrew/bin/codex", "/usr/local/bin/codex"} {
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return path, nil
			}
		}
	}
	return "", i18n.MessageError("codexMissing")
}

// Read owns a bounded app-server session. It never opens credential files.
func Read(ctx context.Context, path string) ([]quota.Window, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "app-server")
	configureProcess(cmd)
	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.New("could not open Codex stdin")
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		input.Close()
		return nil, errors.New("could not open Codex stdout")
	}
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		input.Close()
		return nil, errors.New("could not start Codex")
	}
	defer func() { input.Close(); _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	version, _ := json.Marshal(update.Version)
	if _, err = io.WriteString(input, `{"id":1,"method":"initialize","params":{"clientInfo":{"name":"codex_limit_tray","version":`+string(version)+"}}}\n"); err != nil {
		return nil, errors.New("could not initialize the Codex connection")
	}
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		var message response
		if json.Unmarshal(scanner.Bytes(), &message) != nil {
			continue
		}
		if message.Method == "account/updated" && message.Params.AuthMode == nil {
			return nil, ErrLogin
		}
		if message.ID != 1 && message.ID != 2 {
			continue
		}
		if message.Error != nil {
			return nil, classify(message.Error.Message)
		}
		if message.ID == 2 {
			return quota.Parse(message.Result)
		}
		if _, err = io.WriteString(input, "{\"method\":\"initialized\"}\n{\"id\":2,\"method\":\"account/rateLimits/read\"}\n"); err != nil {
			return nil, errors.New("Codex connection failed")
		}
	}
	if ctx.Err() != nil {
		return nil, errors.New("Codex response timed out")
	}
	return nil, errors.New("Codex connection closed")
}
func classify(message string) error {
	message = strings.ToLower(message)
	for _, word := range []string{"authentication", "unauthorized", "not logged", "401"} {
		if strings.Contains(message, word) {
			return ErrLogin
		}
	}
	if strings.Contains(message, "decode") {
		return ErrVersion
	}
	return i18n.MessageError("codexRead")
}
