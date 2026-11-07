package main

import (
	"context"
	"os"
	"runtime"
	"sync"

	"codexlimittray/internal/ui"
	"codexlimittray/internal/update"
)

// State and decisions (including the automatic-update preference) belong to
// the UI thread. Only HTTPS, hashing and signature verification run off it.
type updateController struct {
	ctx       context.Context
	workers   *sync.WaitGroup
	render    func(ui.UpdateState)
	dispatch  func(func())
	launch    func(string) error
	enabled   func() bool
	installed bool
	exit      func()
	busy      bool
	release   *update.Release
}

func (c *updateController) check(automatic bool) {
	if c.busy || c.ctx.Err() != nil {
		return
	}
	c.busy = true
	c.release = nil
	c.render(ui.UpdateState{Stage: ui.UpdateChecking})
	c.workers.Go(func() {
		release, err := update.Check(c.ctx)
		c.dispatch(func() {
			if c.ctx.Err() != nil {
				return
			}
			c.busy = false
			if err != nil {
				c.render(ui.UpdateState{Stage: ui.UpdateError, Message: err.Error()})
				return
			}
			c.release = release
			state := c.result()
			c.render(state)
			if automatic && c.installed && c.enabled() && release.IsNewer() && state.CanInstall {
				c.install(true)
			}
		})
	})
}

func (c *updateController) result() ui.UpdateState {
	r := c.release
	if r == nil {
		return ui.UpdateState{Stage: ui.UpdateIdle, Message: "Для этой версии пока нет подходящего выпуска на GitHub."}
	}
	s := ui.UpdateState{Stage: ui.UpdateCurrent, Version: r.Version, Notes: r.Notes}
	if !r.IsNewer() && c.installed {
		return s
	}
	if r.IsNewer() {
		s.Stage = ui.UpdateAvailable
	}
	switch {
	case runtime.GOOS != "windows" || runtime.GOARCH != "amd64":
		if r.IsNewer() {
			s.Message = "Новая версия доступна на GitHub. На этой платформе требуется ручная установка."
		}
	case !r.HasInstaller():
		s.Message = "В этом выпуске нет комплекта для проверенной установки Windows."
	case !update.InstallKeyConfigured():
		s.Message = "Сборка без доверенного ключа: установка из приложения отключена."
	default:
		s.CanInstall = true
		if !r.IsNewer() {
			s.Message = "Переносная версия актуальна. Можно установить её для текущего пользователя."
		}
	}
	return s
}

func (c *updateController) install(automatic bool) {
	if c.busy || c.ctx.Err() != nil || !c.result().CanInstall {
		return
	}
	c.busy = true
	state := c.result()
	state.Stage, state.CanInstall = ui.UpdateInstalling, false
	c.render(state)
	download := update.Download
	if !c.installed {
		download = update.DownloadForInstall
	}
	release := c.release
	c.workers.Go(func() {
		path, err := download(c.ctx, release)
		// If the app quits before Fyne dispatches the completion, do not
		// strand the downloaded executable or launch it during shutdown.
		discard := func() {
			if path != "" {
				_ = os.Remove(path)
			}
		}
		stopDiscard := context.AfterFunc(c.ctx, discard)
		c.dispatch(func() {
			if !stopDiscard() || c.ctx.Err() != nil {
				discard()
				return
			}
			c.busy = false
			if automatic && !c.enabled() {
				discard()
				c.render(c.result())
				return
			}
			if err == nil {
				err = c.launch(path)
			}
			if err != nil {
				discard()
				state := c.result()
				state.Stage, state.Message = ui.UpdateError, err.Error()
				c.render(state)
				return
			}
			c.exit()
		})
	})
}
