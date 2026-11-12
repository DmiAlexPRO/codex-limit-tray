package codex

import (
	"bufio"
	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/update"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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

func Read(ctx context.Context, path string) (quota.Limits, error) {
	result, err := call(ctx, path, "account/rateLimits/read", nil)
	if err != nil {
		return quota.Limits{}, err
	}
	return quota.Parse(result)
}

func newResetKey() string {
	var id [16]byte
	rand.Read(id[:])
	id[6] = id[6]&0x0f | 0x40
	id[8] = id[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[:4], id[4:6], id[6:8], id[8:10], id[10:])
}

func consumeReset(ctx context.Context, path, key string) (string, error) {
	result, err := call(ctx, path, "account/rateLimitResetCredit/consume", map[string]string{"idempotencyKey": key})
	if err != nil {
		return "", err
	}
	var reply struct {
		Outcome string `json:"outcome"`
	}
	if json.Unmarshal(result, &reply) == nil {
		switch reply.Outcome {
		case "reset", "alreadyRedeemed":
			return "earnedResetDone", nil
		case "nothingToReset":
			return "earnedResetNotNeeded", nil
		case "noCredit":
			return "earnedResetUnavailable", nil
		}
	}
	return "", errors.New("unsupported reset outcome")
}

// call owns a bounded app-server session. It never opens credential files.
func call(ctx context.Context, path, method string, params map[string]string) (json.RawMessage, error) {
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
	request := struct {
		ID     int               `json:"id"`
		Method string            `json:"method"`
		Params map[string]string `json:"params,omitempty"`
	}{ID: 2, Method: method, Params: params}
	expectedID := 1
	for scanner.Scan() {
		var message response
		if json.Unmarshal(scanner.Bytes(), &message) != nil {
			continue
		}
		if message.Method == "account/updated" && message.Params.AuthMode == nil {
			return nil, ErrLogin
		}
		if message.ID != expectedID {
			continue
		}
		if message.Error != nil {
			return nil, classify(message.Error.Message)
		}
		if message.ID == 2 {
			return message.Result, nil
		}
		if _, err = io.WriteString(input, "{\"method\":\"initialized\"}\n"); err != nil {
			return nil, errors.New("Codex connection failed")
		}
		if err = json.NewEncoder(input).Encode(request); err != nil {
			return nil, errors.New("Codex connection failed")
		}
		expectedID = 2
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
