package ui

import (
	"net/url"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UpdateStage int

const (
	UpdateIdle UpdateStage = iota
	UpdateChecking
	UpdateCurrent
	UpdateAvailable
	UpdateInstalling
	UpdateError
)

// UpdateState is a display snapshot. The caller owns network and installer operations.
type UpdateState struct {
	Stage       UpdateStage
	Version     string
	Notes       string
	Message     string // technical detail, if available
	MessageKey  string
	CanInstall  bool
	RetryAt     time.Time
	ReleasePage *url.URL
}

type updatePanel struct {
	tab             *container.TabItem
	status, version *widget.Label
	notes           *widget.Label
	check, install  *widget.Button
	openRelease     *widget.Button
	automatic       *widget.Check
	installed       bool
	currentVersion  string
	state           UpdateState
}

// ConfigureUpdates adds the settings tab. Nil callbacks mean the corresponding
// action is unavailable (for example in demo mode).
func (v *View) ConfigureUpdates(currentVersion string, installed bool, check, install func()) {
	p := &updatePanel{status: wrapped(""), version: wrapped(""), notes: wrapped(""), installed: installed, currentVersion: currentVersion}
	v.updates = p
	p.check = widget.NewButtonWithIcon(v.tr("checkUpdates"), theme.ViewRefreshIcon(), check)
	p.openRelease = widget.NewButton(v.tr("openReleasePage"), func() {
		if p.state.ReleasePage == nil {
			return
		}
		if err := v.app.OpenURL(p.state.ReleasePage); err != nil {
			dialog.ShowError(err, v.window)
		}
	})
	p.install = highButton(v.tr("installUpdate"), theme.DownloadIcon(), install)
	p.automatic = widget.NewCheck(v.tr("autoUpdate"), nil)
	p.automatic.Checked = installed && v.settings.AutoUpdate
	p.automatic.OnChanged = func(enabled bool) {
		next := v.settings
		next.AutoUpdate = enabled
		if err := v.save(next); err != nil {
			p.automatic.Checked = v.settings.AutoUpdate
			p.automatic.Refresh()
			v.showSaveError(err)
			return
		}
		v.settings = next
	}
	note := v.tr("autoUpdateHint")
	if !installed && runtime.GOOS == "windows" {
		note = v.tr("portableHint")
		p.automatic.Hide()
	}
	if runtime.GOOS != "windows" {
		note = v.tr("manualHint")
	}
	p.tab = container.NewTabItem(v.tr("updates"), scrollPage(container.NewVBox(
		text(v.tr("appUpdates"), 22, true),
		wrapped(v.tr("currentVersion", currentVersion)), p.version, p.status,
		p.install, p.openRelease, p.check, widget.NewSeparator(), p.automatic, wrapped(note),
		p.notes,
	)))
	v.settingsPage.Append(p.tab)
	v.SetUpdateState(UpdateState{Stage: UpdateIdle})
}

// ShowUpdates selects the tab; the caller controls main-window visibility.
func (v *View) ShowUpdates() {
	if v.updates == nil {
		return
	}
	v.ShowSettings()
	v.settingsPage.Select(v.updates.tab)
}

// SetUpdateState must be called on the Fyne UI thread.
func (v *View) SetUpdateState(s UpdateState) {
	p := v.updates
	if p == nil {
		return
	}
	p.state = s
	message := ""
	if !s.RetryAt.IsZero() {
		message = v.tr("updateRateLimited", s.RetryAt.Local().Format("2006-01-02 15:04:05 MST"))
	} else if s.MessageKey != "" {
		message = v.tr(s.MessageKey)
	} else {
		switch s.Stage {
		case UpdateChecking:
			message = v.tr("checking")
		case UpdateCurrent:
			message = v.tr("current")
		case UpdateAvailable:
			message = v.tr("available")
		case UpdateInstalling:
			message = v.tr("installing")
		case UpdateError:
			message = v.tr("updateError")
		default:
			message = v.tr("notChecked")
		}
	}
	if s.Message != "" {
		message += "\n\n" + s.Message
	}
	p.status.SetText(message)
	if s.ReleasePage != nil {
		label := "openReleasePage"
		if s.Stage == UpdateAvailable {
			label = "downloadFromGitHub"
		}
		p.openRelease.SetText(v.tr(label))
		p.openRelease.Show()
	} else {
		p.openRelease.Hide()
	}
	showUpdateText(p.version, v.tr("releaseVersion", s.Version), s.Version != "")
	// GitHub release notes are untrusted. Use a plain label, never rich Markdown or embedded resources.
	showUpdateText(p.notes, v.tr("releaseNotes")+"\n\n"+s.Notes, strings.TrimSpace(s.Notes) != "")
	busy := s.Stage == UpdateChecking || s.Stage == UpdateInstalling
	setEnabled(p.check, !busy && p.check.OnTapped != nil)
	setEnabled(p.automatic, p.installed && p.install.OnTapped != nil)
	if p.installed && ((s.CanInstall && p.install.OnTapped != nil) || s.Stage == UpdateInstalling) {
		p.install.Show()
		setEnabled(p.install, !busy)
	} else {
		p.install.Hide()
	}
}

func showUpdateText(label *widget.Label, value string, visible bool) {
	if !visible {
		label.Hide()
		label.SetText("")
		return
	}
	label.SetText(value)
	label.Show()
}

func setEnabled(control fyne.Disableable, enabled bool) {
	if enabled {
		control.Enable()
		return
	}
	control.Disable()
}
