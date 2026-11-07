package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strconv"
	"sync"

	"codexlimittray/internal/quota"
	"codexlimittray/internal/settings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var fontOnce sync.Once
var indicatorFont *opentype.Font

func Indicator(windows []quota.Window, s settings.Settings) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 128, 128))
	draw.Draw(img, img.Bounds(), image.NewUniform(hex("172030")), image.Point{}, draw.Src)
	if len(windows) == 0 {
		drawNumber(img, "—", 0, 128, 72, hex("95A5BD"))
		return img
	}
	count := min(2, len(windows))
	for index, w := range windows[:count] {
		c := ValueColor(s, w, index)
		y, height := index*128/count, 128/count
		if s.Style == "numbers" {
			size := 57.0
			if count == 1 {
				size = 86
			}
			if w.Remaining == 100 {
				size *= .76
			}
			drawNumber(img, strconv.Itoa(w.Remaining), y, height, size, c)
			continue
		}
		if s.Style == "rings" {
			drawRing(img, 53-index*23, 13, float64(w.Remaining)/100, c)
			continue
		}
		top := y + height/2 - 14
		if s.Style == "segments" {
			for n := 0; n < 10; n++ {
				draw.Draw(img, image.Rect(8+n*11, top, 16+n*11, top+28), image.NewUniform(hex("354359")), image.Point{}, draw.Src)
				fill := int(math.Round(8 * max(0, min(1, float64(w.Remaining)/10-float64(n)))))
				draw.Draw(img, image.Rect(8+n*11, top, 8+n*11+fill, top+28), image.NewUniform(c), image.Point{}, draw.Src)
			}
			continue
		}
		draw.Draw(img, image.Rect(8, top, 120, top+28), image.NewUniform(hex("354359")), image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(8, top, 8+112*w.Remaining/100, top+28), image.NewUniform(c), image.Point{}, draw.Src)
	}
	return img
}
func ValueColor(s settings.Settings, w quota.Window, index int) color.NRGBA {
	if s.ThresholdEnabled && w.Remaining < s.Threshold {
		c, _ := settings.Color(s.LowColor)
		return c
	}
	if w.Minutes == 10080 || (w.Minutes != 300 && index > 0) {
		return hex("B5A1FA")
	}
	c, _ := settings.Color(s.Accent)
	return c
}
func drawNumber(img *image.NRGBA, text string, y, height int, size float64, c color.NRGBA) {
	fontOnce.Do(func() {
		indicatorFont, _ = opentype.Parse(theme.DefaultTheme().Font(fyne.TextStyle{Bold: true}).Content())
	})
	if indicatorFont == nil {
		return
	}
	face, err := opentype.NewFace(indicatorFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return
	}
	defer face.Close()
	width := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()
	baseline := y + (height-metrics.Ascent.Ceil()-metrics.Descent.Ceil())/2 + metrics.Ascent.Ceil()
	d := font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(img.Bounds().Min.X+(img.Bounds().Dx()-width)/2, baseline)}
	d.DrawString(text)
}
func drawRing(img *image.NRGBA, radius, width int, value float64, c color.NRGBA) {
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			dx, dy := float64(x)-63.5, float64(y)-63.5
			r := math.Hypot(dx, dy)
			if r < float64(radius-width/2) || r > float64(radius+width/2) {
				continue
			}
			a := math.Mod(math.Atan2(dy, dx)+math.Pi/2+2*math.Pi, 2*math.Pi)
			pixel := hex("354359")
			if a < value*2*math.Pi {
				pixel = c
			}
			img.SetNRGBA(x, y, pixel)
		}
	}
}
func IconResource(windows []quota.Window, s settings.Settings) fyne.Resource {
	var data bytes.Buffer
	_ = png.Encode(&data, Indicator(windows, s))
	return fyne.NewStaticResource("quota.png", data.Bytes())
}
