package main

import (
	"codexlimittray/internal/i18n"
	"context"
	"flag"
	"runtime"
	"sync"
	"time"

	"codexlimittray/internal/codex"
	"codexlimittray/internal/platform"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"codexlimittray/internal/ui"
	"codexlimittray/internal/update"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"
)

func main() {
	demo := flag.Bool("demo", false, "use clearly labelled sample data, no API access")
	minimized := flag.Bool("minimized", false, "start in the system tray")
	initLanguage := flag.Bool("initialize-language", false, "seed the language for a new Windows installation")
	flag.Parse()
	if *initLanguage {
		if path, err := settings.Path(); err == nil && platform.Installed() && !*demo {
			// Non-fatal: first launch still detects a language if seeding fails.
			_ = settings.InitLanguage(path, i18n.Detect())
		}
		return
	}
	a := app.NewWithID("io.codexlimittray.preview")
	w := platform.NewMainWindow(a, "Codex Limits")
	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()
	show := func() { w.Show(); platform.PrepareMainWindow(w) }
	if !*demo {
		release, primary, err := platform.OwnInstance(func() {
			fyne.Do(func() {
				if ctx.Err() == nil {
					show()
				}
			})
		})
		if err != nil || !primary {
			return
		}
		defer release()
	}
	s := settings.Default()
	path, pathErr := settings.Path()
	var warning error
	if pathErr == nil && !*demo {
		s, warning = settings.Load(path, i18n.Detect())
	}
	if pathErr != nil {
		warning = pathErr
	}
	if !*demo && s.AutoStart && platform.AutoStartSupported() {
		// Refresh the executable location when a portable build has been moved.
		if err := platform.SetAutoStart(true); err != nil {
			warning = err
		}
	}
	refresh := make(chan struct{}, 1)
	var cancel context.CancelFunc
	var workers sync.WaitGroup
	generation := 0
	var view *ui.View
	request := func() {
		select {
		case refresh <- struct{}{}:
		default:
		}
	}
	exit := func() {
		shutdown()
		if cancel != nil {
			cancel()
		}
		a.Quit()
	}
	start := func() {
		if *demo {
			return
		}
		if cancel != nil {
			cancel()
		}
		generation++
		current := generation
		ctx, stop := context.WithCancel(ctx)
		cancel = stop
		resetRequests := make(chan struct{}, 1)
		view.ConfigureReset(func() { resetRequests <- struct{}{} })
		view.UpdateLocalized(nil, func(language string) (string, string) {
			return i18n.Text(language, "connecting"), i18n.Text(language, "waitingReply")
		})
		override := s.CodexPath
		workers.Add(1)
		go func() {
			defer workers.Done()
			codex.Monitor(ctx, override, refresh, resetRequests, func(snapshot codex.Snapshot) {
				fyne.Do(func() {
					if current == generation && ctx.Err() == nil {
						view.SetResetCredits(snapshot.ResetCredits, snapshot.ResetMessage)
						view.UpdateLocalized(snapshot.Windows, func(language string) (string, string) {
							return snapshot.Status(language), snapshot.Diagnostic(language)
						})
					}
				})
			})
		}()
	}
	updateTrayMenu := func(string) {}
	view = ui.NewView(a, w, s, request, exit, func(next settings.Settings) error {
		autoStartChanged := s.AutoStart != next.AutoStart
		if !*demo {
			if pathErr != nil {
				return pathErr
			}
			if autoStartChanged {
				if err := platform.SetAutoStart(next.AutoStart); err != nil {
					return err
				}
			}
			if err := next.Save(path); err != nil {
				if autoStartChanged {
					_ = platform.SetAutoStart(s.AutoStart)
				}
				return err
			}
		}
		languageChanged := s.Language != next.Language
		changed := s.CodexPath != next.CodexPath
		s = next
		if languageChanged {
			updateTrayMenu(s.Language)
		}
		if changed {
			start()
		}
		return nil
	})
	installed := platform.Installed()
	updates := &updateController{ctx: ctx, workers: &workers, render: view.SetUpdateState, installed: installed,
		dispatch: fyne.Do, launch: platform.LaunchInstaller, enabled: func() bool { return s.AutoUpdate }, exit: exit}
	var check, install func()
	switch {
	case *demo:
		view.ConfigureUpdates(update.Version, false, nil, nil)
		view.SetUpdateState(ui.UpdateState{MessageKey: "demoUpdates"})
	case update.Repository == "":
		view.ConfigureUpdates(update.Version, installed, nil, nil)
		view.SetUpdateState(ui.UpdateState{MessageKey: "localBuild"})
	default:
		check = func() { updates.check(false) }
		if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" && update.InstallKeyConfigured() {
			install = func() { updates.install(false) }
		}
		view.ConfigureUpdates(update.Version, installed, check, install)
	}
	if !*demo && runtime.GOOS == "windows" {
		workers.Go(update.Cleanup)
	}
	a.SetIcon(ui.IconResource(nil, s))
	if desk, ok := a.(desktop.App); ok {
		openSettings := func() { view.ShowSettings(); show() }
		updateTrayMenu = trayMenuUpdater(desk.SetSystemTrayMenu, show, request, view.ShowWidget, openSettings, exit)
		updateTrayMenu(s.Language)
		desk.SetSystemTrayIcon(ui.IconResource(nil, s))
		systray.SetOnTapped(func() { fyne.Do(show) })
	}
	if *demo {
		view.UpdateLocalized([]quota.Window{{Remaining: 57, Minutes: 10080, Reset: time.Now().Add(5 * 24 * time.Hour)}}, func(language string) (string, string) {
			return i18n.Text(language, "demo"), i18n.Text(language, "demoOffline")
		})
	} else {
		start()
	}
	if warning != nil {
		view.UpdateLocalized(nil, func(language string) (string, string) {
			return i18n.Text(language, "settingsWarning"), i18n.ErrorText(language, warning)
		})
	}
	if !*minimized {
		show()
	}
	if installed && warning == nil && s.AutoUpdate && install != nil {
		updates.check(true)
	}
	a.Run()
	shutdown()
	if cancel != nil {
		cancel()
	}
	workers.Wait()
}
