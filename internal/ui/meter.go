package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type meter struct {
	widget.BaseWidget
	ratio float32
	color color.Color
}

func newMeter(ratio float32, c color.Color) *meter {
	m := &meter{ratio: ratio, color: c}
	m.ExtendBaseWidget(m)
	return m
}
func (m *meter) CreateRenderer() fyne.WidgetRenderer {
	track := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	track.CornerRadius = 4
	fill := canvas.NewRectangle(m.color)
	fill.CornerRadius = 4
	return &meterRenderer{meter: m, track: track, fill: fill}
}

type meterRenderer struct {
	meter       *meter
	track, fill *canvas.Rectangle
}

func (r *meterRenderer) Layout(size fyne.Size) {
	r.track.Resize(size)
	r.fill.Resize(fyne.NewSize(size.Width*r.meter.ratio, size.Height))
}
func (r *meterRenderer) MinSize() fyne.Size { return fyne.NewSize(80, 8) }
func (r *meterRenderer) Refresh() {
	r.track.FillColor = theme.Color(theme.ColorNameSeparator)
	r.track.Refresh()
	r.fill.Refresh()
}
func (r *meterRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.track, r.fill} }
func (r *meterRenderer) Destroy()                     {}
