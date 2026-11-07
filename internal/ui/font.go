package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func themeFont(bold bool) fyne.Resource {
	return theme.DefaultTheme().Font(fyne.TextStyle{Bold: bold})
}
