package ui

import (
	"runtime"
	"strings"

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
	Stage      UpdateStage
	Version    string
	Notes      string
	Message    string
	CanInstall bool
}

type updatePanel struct {
	tab             *container.TabItem
	status, version *widget.Label
	notes           *widget.Label
	check, install  *widget.Button
	automatic       *widget.Check
	installed       bool
}

// ConfigureUpdates adds the settings tab. Nil callbacks mean the corresponding
// action is unavailable (for example in demo mode).
func (v *View) ConfigureUpdates(currentVersion string, installed bool, check, install func()) {
	p := &updatePanel{status: wrapped(""), version: wrapped(""), notes: wrapped(""), installed: installed}
	v.updates = p
	p.check = widget.NewButtonWithIcon("Проверить обновления", theme.ViewRefreshIcon(), check)
	p.install = highButton("Скачать и установить обновление", theme.DownloadIcon(), install)
	if !installed {
		p.install.SetText("Скачать и установить программу")
	}
	p.automatic = widget.NewCheck("Автоматически устанавливать обновления при запуске", nil)
	p.automatic.Checked = installed && v.settings.AutoUpdate
	p.automatic.OnChanged = func(enabled bool) {
		next := v.settings
		next.AutoUpdate = enabled
		if err := v.save(next); err != nil {
			p.automatic.Checked = v.settings.AutoUpdate
			p.automatic.Refresh()
			dialog.ShowError(err, v.window)
			return
		}
		v.settings = next
	}
	note := "Если настройка включена, при запуске программа проверяет выпуски на GitHub и устанавливает новую версию с перезапуском."
	if !installed && runtime.GOOS == "windows" {
		note = "Переносная версия ничего не устанавливает сама. Кнопка установки запускает установщик. В установленной версии автообновление по умолчанию включено."
	}
	if runtime.GOOS != "windows" {
		note = "Здесь можно проверить версию и прочитать изменения. Устанавливайте новые версии вручную из GitHub Releases."
	}
	p.tab = container.NewTabItem("Обновления", container.NewVScroll(container.NewVBox(
		text("Обновления приложения", 22, true),
		wrapped("Текущая версия: "+currentVersion), p.version, p.status,
		p.install, p.check, widget.NewSeparator(), p.automatic, wrapped(note),
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
	v.showSettings()
	v.settingsPage.Select(v.updates.tab)
}

// SetUpdateState must be called on the Fyne UI thread.
func (v *View) SetUpdateState(s UpdateState) {
	p := v.updates
	if p == nil {
		return
	}
	message := s.Message
	if message == "" {
		switch s.Stage {
		case UpdateChecking:
			message = "Проверяем выпуски на GitHub…"
		case UpdateCurrent:
			message = "У вас актуальная версия."
		case UpdateAvailable:
			message = "Доступна новая версия."
		case UpdateInstalling:
			message = "Подготавливаем установку. Программа перезапустится."
		case UpdateError:
			message = "Не удалось обновить приложение. Попробуйте ещё раз."
		default:
			message = "Проверка ещё не выполнялась."
		}
	}
	p.status.SetText(message)
	showUpdateText(p.version, "Версия выпуска: ", s.Version)
	// GitHub release notes are untrusted. Use a plain label, never rich Markdown or embedded resources.
	showUpdateText(p.notes, "Что изменилось\n\n", s.Notes)
	busy := s.Stage == UpdateChecking || s.Stage == UpdateInstalling
	setEnabled(p.check, !busy && p.check.OnTapped != nil)
	setEnabled(p.automatic, p.installed && p.install.OnTapped != nil)
	if (s.CanInstall && p.install.OnTapped != nil) || s.Stage == UpdateInstalling {
		p.install.Show()
		setEnabled(p.install, !busy)
	} else {
		p.install.Hide()
	}
}

func showUpdateText(label *widget.Label, prefix, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		label.Hide()
		label.SetText("")
		return
	}
	label.SetText(prefix + value)
	label.Show()
}

func setEnabled(control fyne.Disableable, enabled bool) {
	if enabled {
		control.Enable()
		return
	}
	control.Disable()
}
