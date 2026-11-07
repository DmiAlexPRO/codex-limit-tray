package ui

import (
	"fmt"
	"image/color"
	"strings"

	"codexlimittray/internal/platform"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/systray"
)

type View struct {
	app            fyne.App
	window         fyne.Window
	compact        fyne.Window
	floating       *indicatorWidget
	settings       settings.Settings
	windows        []quota.Window
	lastSuccessful []quota.Window
	status         string
	dashboard      *fyne.Container
	diagnostic     *widget.Label
	content        *fyne.Container
	mainPage       fyne.CanvasObject
	settingsPage   *container.AppTabs
	updates        *updatePanel
	navigation     *widget.Button
	onSettings     bool
	refresh        func()
	save           func(settings.Settings) error
}

func NewView(a fyne.App, w fyne.Window, s settings.Settings, refresh, exit func(), save func(settings.Settings) error) *View {
	v := &View{app: a, window: w, settings: s, status: "Ожидание первого запроса", refresh: refresh, save: save}
	a.Settings().SetTheme(appTheme{s})
	v.dashboard = container.NewStack()
	v.diagnostic = wrapped("Данные ещё не запрашивались.")
	v.content = container.NewStack()
	v.navigation = widget.NewButtonWithIcon("", theme.SettingsIcon(), nil)
	v.navigation.Importance = widget.LowImportance
	v.navigation.OnTapped = func() {
		if v.onSettings {
			v.showMain()
		} else {
			v.showSettings()
		}
	}
	controls := container.NewHBox(v.navigation, newWindowControl(theme.WindowMinimizeIcon(), func() { platform.Minimize(w) }), newWindowControl(theme.CancelIcon(), w.Hide))
	header := container.NewBorder(nil, nil, nil, controls, newDragHandle(w, "Codex Limits"))
	header = container.New(layout.NewCustomPaddedLayout(10, 4, 12, 12), header)
	w.SetContent(container.NewPadded(container.NewBorder(header, nil, nil, nil, v.content)))
	w.SetCloseIntercept(w.Hide)
	w.Resize(fyne.NewSize(420, 600))
	v.mainPage = dashboardPage(v.dashboard, v.status, v.refresh, v.ShowWidget)
	v.settingsPage = container.NewAppTabs(
		container.NewTabItem("Оформление", v.preferences()),
		container.NewTabItem("Подключение", v.connection()),
	)
	v.showMain()
	return v
}

func (v *View) showMain() {
	v.onSettings = false
	v.navigation.SetIcon(theme.SettingsIcon())
	v.content.Objects = []fyne.CanvasObject{v.mainPage}
	v.content.Refresh()
}

func (v *View) showSettings() {
	v.onSettings = true
	v.navigation.SetIcon(theme.NavigateBackIcon())
	v.content.Objects = []fyne.CanvasObject{v.settingsPage}
	v.content.Refresh()
}

func (v *View) Update(windows []quota.Window, status, diagnostic string) {
	if v.settings.SoundEnabled && quota.CrossedThreshold(v.lastSuccessful, windows, v.settings.Threshold) {
		platform.Sound()
	}
	if len(windows) > 0 {
		v.lastSuccessful = append([]quota.Window(nil), windows...)
	}
	v.windows = windows
	v.status = status
	v.diagnostic.SetText(diagnostic)
	v.render()
}

func (v *View) render() {
	v.dashboard.Objects = []fyne.CanvasObject{Dashboard(v.windows, v.status, v.settings, v.refresh)}
	v.dashboard.Refresh()
	if desk, ok := v.app.(desktop.App); ok {
		desk.SetSystemTrayIcon(IconResource(v.windows, v.settings))
		systray.SetTooltip(QuotaTooltip(v.windows))
	}
	if v.compact != nil {
		if v.floating == nil {
			v.floating = newIndicatorWidget(v.compact, WidgetIndicator(v.windows, v.settings), v.compact.Hide)
			v.compact.SetContent(v.floating)
		} else {
			v.floating.image = WidgetIndicator(v.windows, v.settings)
			v.floating.Refresh()
		}
	}
}

func dashboardPage(dashboard *fyne.Container, status string, refresh func(), showWidget func()) fyne.CanvasObject {
	actions := container.NewGridWithColumns(2,
		widget.NewButtonWithIcon("Виджет", theme.ViewFullScreenIcon(), showWidget),
		highButton("Обновить лимиты", theme.ViewRefreshIcon(), refresh),
	)
	return container.NewBorder(nil, actions, nil, nil, dashboard)
}

func Dashboard(windows []quota.Window, status string, s settings.Settings, refresh func()) fyne.CanvasObject {
	items := container.NewVBox(
		text("Лимиты", 26, true),
		muted("Остаток по вашей подписке"),
	)
	if len(windows) == 0 {
		items.Add(surface(container.NewVBox(text("Данные не получены", 15, true), wrapped("Состояние и причина указаны ниже. При успешном запросе лимиты появятся автоматически."))))
	}
	for index, w := range windows {
		items.Add(QuotaCard(w, index, s))
	}
	return container.NewBorder(nil, wrapped(status), nil, nil, container.NewVScroll(items))
}

func QuotaCard(w quota.Window, index int, s settings.Settings) fyne.CanvasObject {
	accent := ValueColor(s, w, index)
	value := text(fmt.Sprintf("%d%%", w.Remaining), 30, true)
	progress := newMeter(float32(w.Remaining)/100, accent)
	reset := "Время сброса неизвестно"
	if !w.Reset.IsZero() {
		reset = "Сброс " + w.Reset.Local().Format("02.01 · 15:04")
	}
	marker := canvas.NewRectangle(accent)
	marker.CornerRadius = 2
	marker.SetMinSize(fyne.NewSize(8, 8))
	title := container.NewHBox(container.NewCenter(marker), text(w.Title(), 14, true))
	line := container.NewBorder(nil, nil, title, value, layout.NewSpacer())
	return surface(container.NewVBox(line, progress, muted(reset)))
}

func (v *View) ShowWidget() {
	if v.compact == nil {
		if driver, ok := v.app.Driver().(desktop.Driver); ok {
			v.compact = driver.CreateSplashWindow()
		} else {
			v.compact = v.app.NewWindow("Codex · Виджет")
		}
		v.compact.SetPadded(false)
		v.compact.SetFixedSize(true)
		v.compact.SetCloseIntercept(v.compact.Hide)
		if desk, ok := v.compact.(desktop.Window); ok {
			desk.RequestAlwaysOnTop()
		}
		v.compact.Resize(fyne.NewSize(184, 88))
	}
	v.render()
	v.compact.Show()
	platform.PrepareWindow(v.compact)
}

func highButton(label string, icon fyne.Resource, action func()) *widget.Button {
	button := widget.NewButtonWithIcon(label, icon, action)
	button.Importance = widget.HighImportance
	return button
}

func (v *View) preferences() fyne.CanvasObject {
	themeSelect := widget.NewSelect([]string{"Системная", "Светлая", "Тёмная"}, nil)
	themeValues := []string{"system", "light", "dark"}
	themeSelect.SetSelectedIndex(indexOf(themeValues, v.settings.Theme))
	style := widget.NewSelect([]string{"Цифры", "Бары", "Сегменты", "Кольца"}, nil)
	styles := []string{"numbers", "bars", "segments", "rings"}
	style.SetSelectedIndex(indexOf(styles, v.settings.Style))
	accentValue, lowValue := v.settings.Accent, v.settings.LowColor
	thresholdValue := text(fmt.Sprintf("%d%%", v.settings.Threshold), 14, true)
	threshold := widget.NewSlider(1, 100)
	threshold.Step = 1
	threshold.Value = float64(v.settings.Threshold)
	enabled := widget.NewCheck("Менять цвет ниже порога", nil)
	enabled.Checked = v.settings.ThresholdEnabled
	sound := widget.NewCheck("Звук при переходе ниже порога", nil)
	sound.Checked = v.settings.SoundEnabled
	autoStart := widget.NewCheck("Запускать при входе в систему", nil)
	autoStart.Checked = v.settings.AutoStart
	preview := canvas.NewImageFromImage(Indicator([]quota.Window{{Remaining: 57}, {Remaining: 82}}, v.settings))
	preview.SetMinSize(fyne.NewSize(64, 64))
	preview.FillMode = canvas.ImageFillContain

	draft := func() settings.Settings {
		s := v.settings
		s.Theme = themeValues[max(0, themeSelect.SelectedIndex())]
		s.Style = styles[max(0, style.SelectedIndex())]
		s.Accent = accentValue
		s.LowColor = lowValue
		s.ThresholdEnabled = enabled.Checked
		s.SoundEnabled = sound.Checked
		s.AutoStart = autoStart.Checked
		s.Threshold = int(threshold.Value)
		return s
	}
	apply := func() {
		s := draft()
		if err := s.Validate(); err != nil {
			dialog.ShowError(err, v.window)
			return
		}
		if err := v.save(s); err != nil {
			dialog.ShowError(err, v.window)
			return
		}
		v.settings = s
		v.lastSuccessful = append([]quota.Window(nil), v.windows...)
		v.app.Settings().SetTheme(appTheme{s})
		v.render()
	}
	previewChange := func() {
		s := draft()
		if s.Validate() != nil {
			return
		}
		preview.Image = Indicator([]quota.Window{{Remaining: 57}, {Remaining: 82}}, s)
		preview.Refresh()
	}
	colorButton := func(title string, value *string) fyne.CanvasObject {
		swatch := canvas.NewRectangle(hex(*value))
		swatch.SetMinSize(fyne.NewSize(24, 24))
		button := widget.NewButton("Выбрать…", func() {
			picker := dialog.NewColorPicker(title, "Выберите цвет индикатора", func(c color.Color) {
				if c == nil {
					return
				}
				n := color.NRGBAModel.Convert(c).(color.NRGBA)
				*value = fmt.Sprintf("%02X%02X%02X", n.R, n.G, n.B)
				swatch.FillColor = n
				swatch.Refresh()
				previewChange()
				apply()
			}, v.window)
			picker.Advanced = true
			picker.Show()
			picker.SetColor(hex(*value))
		})
		return container.NewBorder(nil, nil, swatch, nil, button)
	}
	themeSelect.OnChanged = func(string) { apply() }
	style.OnChanged = func(string) { previewChange(); apply() }
	threshold.OnChanged = func(value float64) {
		thresholdValue.Segments[0].(*widget.TextSegment).Text = fmt.Sprintf("%d%%", int(value))
		thresholdValue.Refresh()
		previewChange()
	}
	threshold.OnChangeEnded = func(float64) { apply() }
	enabled.OnChanged = func(bool) { previewChange(); apply() }
	sound.OnChanged = func(bool) { apply() }
	autoStart.OnChanged = func(bool) { apply() }
	form := widget.NewForm(
		widget.NewFormItem("Тема", themeSelect),
		widget.NewFormItem("Значок", style),
		widget.NewFormItem("Основной цвет", colorButton("Основной цвет", &accentValue)),
		widget.NewFormItem("Низкий остаток", colorButton("Низкий остаток", &lowValue)),
		widget.NewFormItem("Порог", container.NewBorder(nil, nil, nil, thresholdValue, threshold)),
	)
	note := wrapped("Изменения сохраняются автоматически. Автозапуск открывает приложение в трее.")
	return container.NewVScroll(container.NewVBox(text("Настройки", 24, true), form, enabled, sound, autoStart, note, container.NewHBox(preview, wrapped("Пример значка: 57 / 82.\nЭто не данные аккаунта."))))
}

func (v *View) connection() fyne.CanvasObject {
	path := widget.NewEntry()
	path.SetPlaceHolder("Автоматический поиск Codex")
	path.SetText(v.settings.CodexPath)
	apply := highButton("Сохранить и проверить", theme.ConfirmIcon(), func() {
		s := v.settings
		s.CodexPath = strings.TrimSpace(path.Text)
		if err := v.save(s); err != nil {
			dialog.ShowError(err, v.window)
			return
		}
		v.settings = s
		v.refresh()
	})
	note := wrapped("Используется существующий вход в Codex. Если требуется авторизация, войдите через Codex desktop или команду codex login. Файлы токенов приложение не читает.")
	return container.NewVScroll(container.NewVBox(text("Подключение", 24, true), note, path, apply, widget.NewSeparator(), v.diagnostic, widget.NewSeparator(), wrapped("Обновление — 30 секунд; при ошибках задержка увеличивается до 5 минут.")))
}

func text(value string, size float32, bold bool) *widget.RichText {
	return widget.NewRichText(&widget.TextSegment{Text: value, Style: widget.RichTextStyle{
		ColorName: theme.ColorNameForeground,
		Inline:    true,
		SizeName:  fyne.ThemeSizeName(fmt.Sprintf("quota.text.%g", size)),
		TextStyle: fyne.TextStyle{Bold: bold},
	}})
}
func muted(value string) *widget.Label {
	label := widget.NewLabel(value)
	label.Importance = widget.LowImportance
	return label
}
func wrapped(value string) *widget.Label {
	label := widget.NewLabel(value)
	label.Wrapping = fyne.TextWrapWord
	return label
}
func surface(content fyne.CanvasObject) fyne.CanvasObject {
	bg := &surfaceBackground{}
	bg.ExtendBaseWidget(bg)
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(12, 12, 12, 12), content))
}
func indexOf(values []string, value string) int {
	for index, item := range values {
		if item == value {
			return index
		}
	}
	return 0
}
