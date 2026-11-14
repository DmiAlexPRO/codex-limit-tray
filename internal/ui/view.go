package ui

import (
	"codexlimittray/internal/i18n"
	"fmt"
	"image/color"
	"net/url"
	"runtime"
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
	language       string
	windows        []quota.Window
	lastSuccessful []quota.Window
	status         string
	statusText     func(string) (string, string)
	dashboard      *fyne.Container
	diagnostic     *widget.Label
	content        *fyne.Container
	mainPage       fyne.CanvasObject
	settingsPage   *container.AppTabs
	updates        *updatePanel
	resets         *resetPanel
	navigation     *widget.Button
	onSettings     bool
	refresh        func()
	save           func(settings.Settings) error
}

func NewView(a fyne.App, w fyne.Window, s settings.Settings, refresh, exit func(), save func(settings.Settings) error) *View {
	tr := func(key string) string { return i18n.Text(s.Language, key) }
	v := &View{app: a, window: w, settings: s, language: s.Language, status: tr("waiting"), refresh: refresh, save: save}
	setWidgetLanguage(s.Language)
	a.Settings().SetTheme(appTheme{s})
	v.resets = newResetPanel(w, v.language)
	v.dashboard = container.NewStack(dashboard(nil, v.status, s, v.resets))
	v.diagnostic = wrapped(tr("diagnosticEmpty"))
	v.content = container.NewStack()
	v.navigation = widget.NewButtonWithIcon("", theme.SettingsIcon(), nil)
	v.navigation.Importance = widget.LowImportance
	v.navigation.OnTapped = func() {
		if v.onSettings {
			v.showMain()
		} else {
			v.ShowSettings()
		}
	}
	controls := container.NewHBox(v.navigation, newWindowControl(theme.WindowMinimizeIcon(), func() { platform.Minimize(w) }), newWindowControl(theme.CancelIcon(), w.Hide))
	header := container.NewBorder(nil, nil, nil, controls, newDragHandle(w, "Codex Limits"))
	header = container.New(layout.NewCustomPaddedLayout(10, 4, 12, 12), header)
	w.SetContent(container.NewPadded(container.NewBorder(header, nil, nil, nil, v.content)))
	w.SetCloseIntercept(w.Hide)
	w.Resize(fyne.NewSize(420, 600))
	v.mainPage = dashboardPage(v.dashboard, v.language, v.refresh, v.ShowWidget)
	v.buildSettingsPage()
	v.showMain()
	return v
}

func (v *View) buildSettingsPage() {
	v.settingsPage = container.NewAppTabs(
		container.NewTabItem(v.tr("appearance"), v.preferences()),
		container.NewTabItem(v.tr("connection"), v.connection()),
	)
}

func (v *View) tr(key string, args ...any) string { return i18n.Text(v.language, key, args...) }

func (v *View) showSaveError(err error) {
	body := wrapped(v.tr("saveError") + "\n\n" + i18n.ErrorText(v.language, err))
	dialog.NewCustom(v.tr("error"), v.tr("ok"), body, v.window).Show()
}

func (v *View) showMain() {
	v.onSettings = false
	v.navigation.SetIcon(theme.SettingsIcon())
	v.content.Objects = []fyne.CanvasObject{v.mainPage}
	v.content.Refresh()
}

// ShowSettings displays the settings page without changing window visibility.
func (v *View) ShowSettings() {
	v.onSettings = true
	v.navigation.SetIcon(theme.NavigateBackIcon())
	v.content.Objects = []fyne.CanvasObject{v.settingsPage}
	v.content.Refresh()
}

func (v *View) Update(windows []quota.Window, status, diagnostic string) {
	v.UpdateLocalized(windows, func(string) (string, string) { return status, diagnostic })
}

// UpdateLocalized keeps the latest status available in the next selected language.
func (v *View) UpdateLocalized(windows []quota.Window, statusText func(string) (string, string)) {
	if v.settings.SoundEnabled && quota.CrossedThreshold(v.lastSuccessful, windows, v.settings.Threshold) {
		platform.Sound()
	}
	if len(windows) > 0 {
		v.lastSuccessful = append([]quota.Window(nil), windows...)
	}
	v.windows = windows
	v.statusText = statusText
	v.localizeStatus()
	v.render()
}

func (v *View) localizeStatus() {
	if v.statusText == nil {
		v.status = v.tr("waiting")
		v.diagnostic.SetText(v.tr("diagnosticEmpty"))
		return
	}
	status, diagnostic := v.statusText(v.language)
	v.status = status
	v.diagnostic.SetText(diagnostic)
}

func (v *View) render() {
	display := v.settings
	display.Language = v.language
	v.dashboard.Objects = []fyne.CanvasObject{dashboard(v.windows, v.status, display, v.resets)}
	v.dashboard.Refresh()
	if desk, ok := v.app.(desktop.App); ok {
		desk.SetSystemTrayIcon(IconResource(v.windows, v.settings))
		systray.SetTooltip(QuotaTooltip(v.windows, v.language))
	}
	if v.compact != nil {
		if v.floating == nil {
			v.floating = newIndicatorWidget(v.compact, WidgetIndicator(v.windows, display), v.compact.Hide)
			v.compact.SetContent(v.floating)
		} else {
			v.floating.image = WidgetIndicator(v.windows, display)
			v.floating.Refresh()
		}
	}
}

func dashboardPage(dashboard *fyne.Container, language string, refresh func(), showWidget func()) fyne.CanvasObject {
	widgetButton := widget.NewButtonWithIcon(i18n.Text(language, "widget"), theme.ViewFullScreenIcon(), showWidget)
	refreshButton := highButton(i18n.Text(language, "refresh"), theme.ViewRefreshIcon(), refresh)
	actions := container.NewBorder(nil, nil, widgetButton, nil, refreshButton)
	return container.NewBorder(nil, actions, nil, nil, dashboard)
}

func dashboard(windows []quota.Window, status string, s settings.Settings, resets *resetPanel) fyne.CanvasObject {
	tr := func(key string) string { return i18n.Text(s.Language, key) }
	items := container.NewVBox(
		text(tr("limits"), 26, true),
		wrapped(tr("remaining")),
	)
	if len(windows) == 0 {
		items.Add(surface(container.NewVBox(text(tr("noData"), 15, true), wrapped(tr("noDataHint")))))
	}
	for index, w := range windows {
		items.Add(QuotaCard(w, index, s))
	}
	items.Add(resets.body)
	scroll := scrollPage(items)
	resets.relayout = func() {
		scroll.Refresh()
		items.Refresh()
	}
	return container.NewBorder(nil, wrapped(status), nil, nil, scroll)
}

func QuotaCard(w quota.Window, index int, s settings.Settings) fyne.CanvasObject {
	accent := ValueColor(s, w, index)
	value := text(fmt.Sprintf("%d%%", w.Remaining), 30, true)
	progress := newMeter(float32(w.Remaining)/100, accent)
	reset := i18n.Text(s.Language, "resetUnknown")
	if !w.Reset.IsZero() {
		reset = i18n.Text(s.Language, "reset", w.Reset.Local().Format("02.01 · 15:04"))
	}
	marker := canvas.NewRectangle(accent)
	marker.CornerRadius = 2
	marker.SetMinSize(fyne.NewSize(8, 8))
	title := container.NewHBox(container.NewCenter(marker), text(quotaTitle(w, s.Language), 14, true))
	line := container.NewBorder(nil, nil, title, value, layout.NewSpacer())
	return surface(container.NewVBox(line, progress, wrapped(reset)))
}

func (v *View) ShowWidget() {
	if v.compact == nil {
		if driver, ok := v.app.Driver().(desktop.Driver); ok {
			v.compact = driver.CreateSplashWindow()
		} else {
			v.compact = v.app.NewWindow("Codex · " + v.tr("widget"))
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
	tr := func(key string) string { return v.tr(key) }
	language := v.languageSelector()

	themeValues := []string{"system", "light", "dark"}
	styles := []string{"numbers", "bars", "segments", "rings"}
	themeSelect := localizedSelect(tr("theme"), []string{tr("system"), tr("light"), tr("dark")}, indexOf(themeValues, v.settings.Theme))
	style := localizedSelect(tr("icon"), []string{tr("numbers"), tr("bars"), tr("segments"), tr("rings")}, indexOf(styles, v.settings.Style))
	accentValue, lowValue := v.settings.Accent, v.settings.LowColor
	thresholdValue := text(fmt.Sprintf("%d%%", v.settings.Threshold), 14, true)
	threshold := widget.NewSlider(1, 100)
	threshold.Step = 1
	threshold.Value = float64(v.settings.Threshold)
	enabled := widget.NewCheck(tr("colorBelow"), nil)
	enabled.Checked = v.settings.ThresholdEnabled
	sound := widget.NewCheck(tr("soundBelow"), nil)
	sound.Checked = v.settings.SoundEnabled
	autoStart := widget.NewCheck(tr("autostart"), nil)
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
			v.showSaveError(err)
			return
		}
		if err := v.save(s); err != nil {
			v.showSaveError(err)
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
		button := widget.NewButton(title+"…", func() {
			picker := dialog.NewColorPicker(title, tr("chooseColor"), func(c color.Color) {
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
		widget.NewFormItem(tr("language"), language),
		widget.NewFormItem(tr("theme"), themeSelect),
		widget.NewFormItem(tr("icon"), style),
		widget.NewFormItem(tr("threshold"), container.NewBorder(nil, nil, nil, thresholdValue, threshold)),
	)
	note := wrapped(tr("autosave"))
	content := container.NewVBox(text(tr("settings"), 24, true), form,
		colorButton(tr("accent"), &accentValue), colorButton(tr("lowColor"), &lowValue),
		enabled, sound, autoStart, note, container.NewBorder(nil, nil, preview, nil, wrapped(tr("preview"))))
	for _, item := range v.trayVisibilityHelp(runtime.GOOS) {
		content.Add(item)
	}
	return scrollPage(content)
}

func (v *View) trayVisibilityHelp(goos string) []fyne.CanvasObject {
	switch goos {
	case "windows":
		openSettings := widget.NewButton(v.tr("openTaskbarSettings"), func() {
			if err := v.app.OpenURL(&url.URL{Scheme: "ms-settings", Opaque: "taskbar"}); err != nil {
				body := wrapped(v.tr("openTaskbarSettingsError"))
				dialog.NewCustom(v.tr("error"), v.tr("ok"), body, v.window).Show()
			}
		})
		return []fyne.CanvasObject{widget.NewSeparator(), wrapped(v.tr("taskbarIconHelp")), openSettings}
	case "darwin":
		return []fyne.CanvasObject{widget.NewSeparator(), wrapped(v.tr("menuBarIconHelp"))}
	default:
		return nil
	}
}

func (v *View) languageSelector() *widget.Select {
	languages := i18n.Languages()
	names := make([]string, 0, len(languages))
	codes := make([]string, 0, len(languages))
	for _, language := range languages {
		names = append(names, language.Name)
		codes = append(codes, language.Code)
	}
	selectLanguage := localizedSelect(v.tr("language"), names, indexOf(codes, v.settings.Language))
	selectLanguage.OnChanged = func(name string) {
		next := v.settings
		next.Language = codes[indexOf(names, name)]
		if next.Language == v.settings.Language {
			return
		}
		if err := v.save(next); err != nil {
			selectLanguage.Selected = names[indexOf(codes, v.settings.Language)]
			selectLanguage.Refresh()
			v.showSaveError(err)
			return
		}
		v.applyLanguage(next)
	}
	return selectLanguage
}

func (v *View) applyLanguage(next settings.Settings) {
	v.settings = next
	v.language = next.Language
	setWidgetLanguage(v.language)
	v.localizeStatus()
	v.resets.language = next.Language
	v.resets.button.SetText(v.tr("earnedResetUse"))
	v.resets.render()
	v.mainPage = dashboardPage(v.dashboard, v.language, v.refresh, v.ShowWidget)
	if v.compact != nil {
		v.compact.SetTitle("Codex · " + v.tr("widget"))
	}
	selected := v.settingsPage.SelectedIndex()
	v.buildSettingsPage()
	if p := v.updates; p != nil {
		v.ConfigureUpdates(p.currentVersion, p.installed, p.check.OnTapped, p.install.OnTapped)
		v.SetUpdateState(p.state)
	}
	v.settingsPage.SelectIndex(selected)
	v.render()
	if v.onSettings {
		v.ShowSettings()
		return
	}
	v.showMain()
}

func (v *View) connection() fyne.CanvasObject {
	tr := func(key string) string { return v.tr(key) }
	path := widget.NewEntry()
	path.SetPlaceHolder(tr("findCodex"))
	path.SetText(v.settings.CodexPath)
	apply := highButton(tr("saveCheck"), theme.ConfirmIcon(), func() {
		s := v.settings
		s.CodexPath = strings.TrimSpace(path.Text)
		if s.CodexPath != v.settings.CodexPath && v.resets.pending {
			v.showSaveError(i18n.MessageError("earnedResetPending"))
			return
		}
		if err := v.save(s); err != nil {
			v.showSaveError(err)
			return
		}
		v.settings = s
		v.refresh()
	})
	note := wrapped(tr("signin"))
	return scrollPage(container.NewVBox(text(tr("connection"), 24, true), note, path, apply, widget.NewSeparator(), v.diagnostic, widget.NewSeparator(), wrapped(tr("polling"))))
}

// Fyne draws the scrollbar over the content, so leave room for its expanded width.
func scrollPage(content fyne.CanvasObject) *container.Scroll {
	return container.NewVScroll(container.New(layout.NewCustomPaddedLayout(0, 0, 0, theme.ScrollBarSize()), content))
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

func localizedSelect(label string, options []string, selected int) *widget.Select {
	control := &widget.Select{Options: options, Selected: options[selected], PlaceHolder: label}
	control.ExtendBaseWidget(control)
	return control
}
