package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "app-server" && os.Getenv("CODEX_LIMIT_TEST_SERVER") != "" {
		fakeServer()
		return
	}
	os.Exit(m.Run())
}

func fakeServer() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request struct {
			Method string `json:"method"`
		}
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			os.Exit(2)
		}
		switch request.Method {
		case "initialize":
			fmt.Println(`{"id":1,"result":{}}`)
		case "initialized":
		case "account/rateLimits/read":
			switch os.Getenv("CODEX_LIMIT_TEST_SERVER") {
			case "success":
				fmt.Println(`{"id":2,"result":{"rateLimits":{"primary":{"usedPercent":43,"windowDurationMins":10080}}}}`)
			case "login":
				fmt.Println(`{"id":2,"error":{"message":"401 Unauthorized secret"}}`)
			case "invalid":
				fmt.Println(`{"id":2,"result":{"rateLimits":null}}`)
			case "timeout":
				time.Sleep(10 * time.Second)
			}
			return
		default:
			os.Exit(2)
		}
	}
}

func TestAppServerSession(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"success", "login", "invalid", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("CODEX_LIMIT_TEST_SERVER", mode)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			windows, err := Read(ctx, path)
			if mode == "success" {
				if err != nil || len(windows) != 1 || windows[0].Remaining != 57 {
					t.Fatalf("%v %v", windows, err)
				}
				return
			}
			if err == nil || len(windows) != 0 {
				t.Fatalf("failed request returned values: %v %v", windows, err)
			}
			if mode == "login" && !errors.Is(err, ErrLogin) {
				t.Fatal(err)
			}
		})
	}
}

func TestSafeErrors(t *testing.T) {
	if !errors.Is(classify("401 Unauthorized secret"), ErrLogin) {
		t.Fatal("auth not classified")
	}
	if !errors.Is(classify("Decode error secret"), ErrVersion) {
		t.Fatal("decode not classified")
	}
	if classify("secret").Error() == "secret" {
		t.Fatal("raw server data exposed")
	}
}
func TestRuntimeOverride(t *testing.T) {
	if _, err := Resolve("relative.exe"); err == nil {
		t.Fatal("relative runtime accepted")
	}
	path := filepath.Join(t.TempDir(), "codex.exe")
	if err := os.WriteFile(path, nil, 0700); err != nil {
		t.Fatal(err)
	}
	if got, err := Resolve(path); err != nil || got != path {
		t.Fatalf("%s %v", got, err)
	}
}
func TestMonitorCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { Monitor(ctx, "", nil, func(Snapshot) { t.Error("cancelled monitor published") }); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancel stalled")
	}
}
