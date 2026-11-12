//go:build windows && amd64

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"codexlimittray/internal/settings"
	"codexlimittray/internal/ui"
	"codexlimittray/internal/update"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

const installerFixture = "inert installer fixture; never execute"

type releaseTransport func(*http.Request) (*http.Response, error)

func (rt releaseTransport) RoundTrip(req *http.Request) (*http.Response, error) { return rt(req) }

// Only the wire and the final OS launch are faked. Keep global build variables
// and the default transport isolated in this test process; never run in parallel.
func fakeRelease(t *testing.T, version string, badSignature bool) *int {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	oldVersion, oldRepo, oldKey, oldTransport := update.Version, update.Repository, update.PublicKey, http.DefaultTransport
	t.Cleanup(func() {
		update.Version, update.Repository, update.PublicKey, http.DefaultTransport = oldVersion, oldRepo, oldKey, oldTransport
	})
	update.Version, update.Repository, update.PublicKey = "1.0.0", "example/preview", base64.StdEncoding.EncodeToString(pub)
	message := update.SignatureMessage(version, fmt.Sprintf("%x", sha256.Sum256([]byte(installerFixture))))
	sig := ed25519.Sign(key, message)
	if badSignature {
		sig[0] ^= 1
	}
	asset := "https://github.com/example/preview/releases/download/v" + version + "/" + update.AssetName
	metadata := fmt.Sprintf(`{"tag_name":"v%s","body":"Release notes","assets":[{"name":"%s"},{"name":"%s.sig"}]}`, version, update.AssetName, update.AssetName)
	requests := 0
	http.DefaultTransport = releaseTransport(func(req *http.Request) (*http.Response, error) {
		requests++
		var data string
		switch req.URL.String() {
		case "https://api.github.com/repos/example/preview/releases/latest":
			data = metadata
		case asset:
			data = installerFixture
		case asset + ".sig":
			data = base64.StdEncoding.EncodeToString(sig)
		default:
			return nil, fmt.Errorf("unexpected fixture request: %s", req.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(data)), ContentLength: int64(len(data))}, nil
	})
	return &requests
}

// Search visible controls in the active tab, so a hidden installer or a disabled
// button cannot accidentally count as the user's explicit installation action.
func updateControl(o fyne.CanvasObject, label string) fyne.Tappable {
	if !o.Visible() {
		return nil
	}
	switch o := o.(type) {
	case *widget.Button:
		if o.Text == label && !o.Disabled() {
			return o
		}
	case *widget.Check:
		if o.Text == label && !o.Disabled() {
			return o
		}
	case *container.AppTabs:
		return updateControl(o.Selected().Content, label)
	case *container.Scroll:
		return updateControl(o.Content, label)
	case *fyne.Container:
		for _, child := range o.Objects {
			if match := updateControl(child, label); match != nil {
				return match
			}
		}
	}
	return nil
}

func TestUpdatesThroughSignedDownload(t *testing.T) {
	for _, scenario := range []struct {
		name, version                  string
		portable, optOut, badSignature bool
		wantLaunch                     bool
	}{
		{name: "automatic", version: "1.1.0", wantLaunch: true},
		{name: "opt out before launch", version: "1.1.0", optOut: true},
		{name: "invalid signature", version: "1.1.0", badSignature: true},
		{name: "portable exact version via button", version: "1.0.0", portable: true, wantLaunch: true},
		{name: "already current", version: "1.0.0"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("LocalAppData", tmp)
			t.Setenv("TMP", tmp)
			t.Setenv("TEMP", tmp)
			requests := fakeRelease(t, scenario.version, scenario.badSignature)
			a := test.NewApp()
			defer a.Quit()
			w := a.NewWindow("update flow")
			s := settings.Default()
			view := ui.NewView(a, w, s, func() {}, func() {}, func(next settings.Settings) error { s = next; return nil })
			last := ui.UpdateState{}
			launches, exits := 0, 0
			var workers sync.WaitGroup
			defer workers.Wait()
			queue := make(chan func(), 1)
			c := updateController{ctx: t.Context(), workers: &workers, installed: !scenario.portable,
				render:  func(state ui.UpdateState) { last = state; view.SetUpdateState(state) },
				enabled: func() bool { return s.AutoUpdate }, exit: func() { exits++ },
				dispatch: func(f func()) { queue <- f },
				launch: func(path string) error {
					launches++
					got, err := os.ReadFile(path)
					if err != nil || string(got) != installerFixture {
						t.Fatalf("launcher got invalid payload %q: %v", got, err)
					}
					return os.Remove(path)
				},
			}
			view.ConfigureUpdates(update.Version, c.installed, func() { c.check(false) }, func() { c.install(false) })
			view.ShowUpdates()
			w.Show()
			tap := func(label string) {
				t.Helper()
				control := updateControl(w.Content(), label)
				if control == nil {
					t.Fatalf("no available action %q", label)
				}
				test.Tap(control)
			}
			complete := func() {
				t.Helper()
				workers.Wait()
				select {
				case f := <-queue:
					f()
				default:
					t.Fatal("worker did not dispatch its result")
				}
			}
			if scenario.portable {
				tap("Check for updates")
			} else {
				c.check(true)
			}
			complete()
			if scenario.portable {
				if c.busy || *requests != 1 || launches != 0 || last.Stage != ui.UpdateCurrent || last.Notes != "Release notes" {
					t.Fatalf("portable updated without consent or lost notes: %+v", last)
				}
				tap("Install app")
			}
			if !scenario.portable && scenario.version == update.Version {
				if c.busy || last.Stage != ui.UpdateCurrent || last.CanInstall || last.Notes != "Release notes" || *requests != 1 {
					t.Fatalf("unexpected reinstallation of current release: %+v", last)
				}
			} else {
				workers.Wait() // verified download has completed; OS launch still awaits UI dispatch
				if scenario.optOut {
					tap("Auto-install updates at startup")
				}
				complete()
				if *requests != 3 || scenario.badSignature && (last.Stage != ui.UpdateError || !strings.Contains(last.Message, "signature")) {
					t.Fatalf("unexpected download outcome: requests=%d, state=%+v", *requests, last)
				}
			}
			if (launches == 1) != scenario.wantLaunch || exits != launches || launches > 1 || c.busy {
				t.Fatalf("unexpected launch/exit: %d/%d; busy=%v; state=%+v", launches, exits, c.busy, last)
			}
			files, err := filepath.Glob(filepath.Join(tmp, "CodexLimitTray", "updates", "*.exe"))
			if err != nil || len(files) != 0 {
				t.Fatalf("orphaned installer: %v %v", files, err)
			}
		})
	}
}

func TestUpdateChecksWaitForRateLimit(t *testing.T) {
	fakeRelease(t, "1.1.0", false)
	requests := 0
	http.DefaultTransport = releaseTransport(func(*http.Request) (*http.Response, error) {
		requests++
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Header:     http.Header{"Retry-After": []string{"120"}},
			Body:       io.NopCloser(strings.NewReader("rate limit reached")),
		}, nil
	})
	var workers sync.WaitGroup
	var state ui.UpdateState
	dispatch := make(chan func(), 1)
	c := &updateController{
		ctx: t.Context(), workers: &workers,
		render:   func(next ui.UpdateState) { state = next },
		dispatch: func(fn func()) { dispatch <- fn },
	}
	c.check(false)
	workers.Wait()
	(<-dispatch)()
	if state.Stage != ui.UpdateError || !state.RetryAt.After(time.Now()) || state.Message != "" || state.ReleasePage == nil {
		t.Fatalf("expected actionable rate-limit state: %+v", state)
	}
	c.check(false)
	if requests != 1 || c.busy {
		t.Fatal("checked GitHub before the retry time")
	}
	c.rateLimit.RetryAt = time.Now().Add(-time.Second)
	c.check(false)
	workers.Wait()
	(<-dispatch)()
	if requests != 2 {
		t.Fatal("did not resume checks after the retry time")
	}
}
