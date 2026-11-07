package ui

import (
	"codexlimittray/internal/settings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"math"
	"strconv"
	"strings"
)

type appTheme struct{ settings settings.Settings }

func (t appTheme) variant(v fyne.ThemeVariant) fyne.ThemeVariant {
	if t.settings.Theme == "dark" {
		return theme.VariantDark
	}
	if t.settings.Theme == "light" {
		return theme.VariantLight
	}
	return v
}
func (t appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	variant = t.variant(variant)
	if name == theme.ColorNamePrimary || name == theme.ColorNameForegroundOnPrimary {
		c, _ := settings.Color(t.settings.Accent)
		if variant == theme.VariantLight {
			c.R, c.G, c.B = uint8(float64(c.R)*0.72), uint8(float64(c.G)*0.72), uint8(float64(c.B)*0.72)
		}
		if name == theme.ColorNameForegroundOnPrimary {
			return contrastText(c)
		}
		return c
	}
	if name == theme.ColorNameFocus || name == theme.ColorNameSelection {
		if variant == theme.VariantDark {
			return hex("29493F")
		}
		return hex("DCEFE9")
	}
	if variant == theme.VariantDark {
		switch name {
		case theme.ColorNameBackground:
			return hex("0E141C")
		case theme.ColorNameInputBackground:
			return hex("151D27")
		case theme.ColorNameButton, theme.ColorNameMenuBackground:
			return hex("18222D")
		case theme.ColorNameForeground:
			return hex("F0F4F8")
		case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
			return hex("8D9AA8")
		case theme.ColorNameSeparator:
			return hex("2A3440")
		case theme.ColorNameShadow:
			return hex("05080C")
		}
	}
	switch name {
	case theme.ColorNameBackground:
		return hex("F4F7F9")
	case theme.ColorNameInputBackground:
		return hex("FFFFFF")
	case theme.ColorNameButton, theme.ColorNameMenuBackground:
		return hex("E9EEF2")
	case theme.ColorNameForeground:
		return hex("17232E")
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return hex("687784")
	case theme.ColorNameSeparator:
		return hex("DFE6EB")
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Pick the higher-contrast of black and white for the user-selected sRGB accent.
func contrastText(c color.NRGBA) color.Color {
	linear := func(channel uint8) float64 {
		n := float64(channel) / 255
		if n <= 0.04045 {
			return n / 12.92
		}
		return math.Pow((n+0.055)/1.055, 2.4)
	}
	luminance := 0.2126*linear(c.R) + 0.7152*linear(c.G) + 0.0722*linear(c.B)
	if (luminance+0.05)/0.05 > 1.05/(luminance+0.05) {
		return color.Black
	}
	return color.White
}
func (t appTheme) Size(n fyne.ThemeSizeName) float32 {
	if value, ok := strings.CutPrefix(string(n), "quota.text."); ok {
		size, err := strconv.ParseFloat(value, 32)
		if err == nil && size > 0 && size <= 100 {
			return float32(size)
		}
	}
	switch n {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameText:
		return 12
	case theme.SizeNameInputRadius, theme.SizeNameButtonRadius:
		return 10
	case theme.SizeNameInnerPadding:
		return 10
	}
	return theme.DefaultTheme().Size(n)
}
func (t appTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (t appTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func hex(s string) color.NRGBA                             { c, _ := settings.Color(s); return c }

type surfaceBackground struct {
	widget.BaseWidget
	rect *canvas.Rectangle
}

func (b *surfaceBackground) CreateRenderer() fyne.WidgetRenderer {
	b.rect = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	b.rect.CornerRadius = 14
	b.rect.StrokeColor = theme.Color(theme.ColorNameSeparator)
	b.rect.StrokeWidth = 1
	return widget.NewSimpleRenderer(b.rect)
}

func (b *surfaceBackground) Refresh() {
	if b.rect != nil {
		b.rect.FillColor = theme.Color(theme.ColorNameInputBackground)
		b.rect.StrokeColor = theme.Color(theme.ColorNameSeparator)
	}
	b.BaseWidget.Refresh()
}
