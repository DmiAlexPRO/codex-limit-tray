package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
			Params struct {
				Key string `json:"idempotencyKey"`
			} `json:"params"`
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
			case "redeem":
				history, _ := os.ReadFile(os.Getenv("CODEX_LIMIT_TEST_HISTORY"))
				count := 2
				if strings.Count(string(history), "\n") == 2 {
					count = 1
				}
				fmt.Printf(`{"id":2,"result":{"rateLimits":{"primary":{"usedPercent":43}},"rateLimitResetCredits":{"availableCount":%d}}}`+"\n", count)
			}
			return
		case "account/rateLimitResetCredit/consume":
			if request.Params.Key == "" {
				os.Exit(2)
			}
			mode := os.Getenv("CODEX_LIMIT_TEST_SERVER")
			if mode != "redeem" {
				fmt.Printf(`{"id":2,"result":{"outcome":%q}}`+"\n", mode)
				return
			}
			history := os.Getenv("CODEX_LIMIT_TEST_HISTORY")
			first, _ := os.ReadFile(history)
			if os.WriteFile(history, append(first, (request.Params.Key+"\n")...), 0600) != nil {
				os.Exit(2)
			}
			if len(first) != 0 {
				fmt.Println(`{"id":2,"result":{"outcome":"alreadyRedeemed"}}`)
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
			limits, err := Read(ctx, path)
			windows := limits.Windows
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

func TestResetOutcomes(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for outcome, want := range map[string]string{
		"reset": "earnedResetDone", "alreadyRedeemed": "earnedResetDone", "noCredit": "earnedResetUnavailable",
		"nothingToReset": "earnedResetNotNeeded", "unknown": "",
	} {
		t.Run(outcome, func(t *testing.T) {
			t.Setenv("CODEX_LIMIT_TEST_SERVER", outcome)
			got, err := consumeReset(t.Context(), path, newResetKey())
			if got != want || (err == nil) != (want != "") {
				t.Fatalf("reset outcome: %q, %v", got, err)
			}
		})
	}
}

func TestMonitorResetRetryRefreshesCredits(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	history := filepath.Join(t.TempDir(), "requests")
	t.Setenv("CODEX_LIMIT_TEST_SERVER", "redeem")
	t.Setenv("CODEX_LIMIT_TEST_HISTORY", history)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	redeem := make(chan struct{}, 1)
	updates := make(chan Snapshot, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		Monitor(ctx, path, nil, redeem, func(s Snapshot) { updates <- s })
	}()
	defer func() { cancel(); <-done }()
	check := func(message string, count int) {
		t.Helper()
		select {
		case snapshot := <-updates:
			if snapshot.Error != nil || snapshot.ResetMessage != message || snapshot.ResetCredits == nil || snapshot.ResetCredits.Available != count {
				t.Fatalf("unexpected snapshot: %+v", snapshot)
			}
		case <-ctx.Done():
			t.Fatal("monitor did not publish")
		}
	}
	check("", 2)
	redeem <- struct{}{}
	check("earnedResetUnconfirmed", 2)
	redeem <- struct{}{}
	check("earnedResetDone", 1)
	requests, err := os.ReadFile(history)
	keys := strings.Fields(string(requests))
	if err != nil || len(keys) != 2 || keys[0] != keys[1] ||
		!regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`).MatchString(keys[0]) {
		t.Fatalf("expected the same UUID across both requests, got %q (%v)", keys, err)
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
	go func() {
		Monitor(ctx, "", nil, nil, func(Snapshot) { t.Error("cancelled monitor published") })
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancel stalled")
	}
}
