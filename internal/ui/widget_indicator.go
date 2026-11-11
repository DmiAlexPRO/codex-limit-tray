package ui

import (
	"codexlimittray/internal/i18n"
	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	stdDraw "image/draw"
	"strings"
)

func WidgetIndicator(windows []quota.Window, s settings.Settings) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 368, 176))
	stdDraw.Draw(img, img.Bounds(), image.NewUniform(hex("111923")), image.Point{}, stdDraw.Src)
	if len(windows) == 0 {
		drawWidgetText(img, "Codex Limits", 30, 56, 23, hex("EDF3F8"), false, true)
		drawWidgetText(img, "—", 30, 112, 21, hex("8D9AA8"), false, false)
		return img
	}
	count := min(2, len(windows))
	for index, w := range windows[:count] {
		y := 64 + index*70
		if count == 1 {
			y = 103
		}
		label := "5h"
		if w.Minutes == 10080 {
			label = "7d"
		} else if w.Minutes != 300 {
			label = "—"
			if w.Minutes > 0 {
				label = fmt.Sprintf("%dm", w.Minutes)
			}
		}
		drawWidgetText(img, label, 28, y, 20, hex("8996A4"), false, false)
		accent := ValueColor(s, w, index)
		drawWidgetText(img, fmt.Sprintf("%d%%", w.Remaining), 340, y, 32, accent, true, true)
		track := image.Rect(94, y-17, 236, y-9)
		stdDraw.Draw(img, track, image.NewUniform(hex("2A3541")), image.Point{}, stdDraw.Src)
		fill := track
		fill.Max.X = fill.Min.X + track.Dx()*w.Remaining/100
		stdDraw.Draw(img, fill, image.NewUniform(accent), image.Point{}, stdDraw.Src)
	}
	return img
}

func drawWidgetText(img *image.NRGBA, value string, x, y, size int, c color.NRGBA, right, bold bool) {
	fontOnce.Do(func() {
		indicatorFont, _ = opentype.Parse(themeFont(bold).Content())
	})
	if indicatorFont == nil {
		return
	}
	face, err := opentype.NewFace(indicatorFont, &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return
	}
	defer face.Close()
	if right {
		x -= font.MeasureString(face, value).Ceil()
	}
	d := font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
	d.DrawString(value)
}

func quotaTitle(w quota.Window, language string) string {
	switch w.Minutes {
	case 300:
		return i18n.Text(language, "fiveHours")
	case 10080:
		return i18n.Text(language, "week")
	case 0:
		return i18n.Text(language, "limit")
	default:
		return i18n.Text(language, "minutes", w.Minutes)
	}
}

func QuotaTooltip(windows []quota.Window, language string) string {
	if len(windows) == 0 {
		return "Codex Limits — " + i18n.Text(language, "noData")
	}
	lines := []string{"Codex Limits"}
	for _, w := range windows {
		lines = append(lines, fmt.Sprintf("%s: %d%%", quotaTitle(w, language), w.Remaining))
	}
	return strings.Join(lines, "\n")
}
