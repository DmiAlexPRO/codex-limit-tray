package ui

import (
	"codexlimittray/internal/platform"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image"
)

type dragHandle struct {
	widget.BaseWidget
	window fyne.Window
	label  string
}

func newDragHandle(w fyne.Window, label string) *dragHandle {
	h := &dragHandle{window: w, label: label}
	h.ExtendBaseWidget(h)
	return h
}
func (h *dragHandle) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(widget.NewLabel(h.label))
}
func (h *dragHandle) Dragged(e *fyne.DragEvent) { platform.Drag(h.window, e.Dragged.DX, e.Dragged.DY) }
func (h *dragHandle) DragEnd()                  {}

type indicatorWidget struct {
	widget.BaseWidget
	window fyne.Window
	image  image.Image
	hide   func()
	hover  bool
}

func newIndicatorWidget(w fyne.Window, img image.Image, hide func()) *indicatorWidget {
	i := &indicatorWidget{window: w, image: img, hide: hide}
	i.ExtendBaseWidget(i)
	return i
}
func (i *indicatorWidget) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromImage(i.image)
	img.SetMinSize(fyne.NewSize(184, 88))
	img.FillMode = canvas.ImageFillContain
	circle := canvas.NewCircle(hex("455268"))
	circle.Resize(fyne.NewSize(18, 18))
	circle.Move(fyne.NewPos(10, 8))
	icon := canvas.NewImageFromResource(theme.CancelIcon())
	icon.Resize(fyne.NewSize(12, 12))
	icon.Move(fyne.NewPos(13, 11))
	if !i.hover {
		circle.Hide()
		icon.Hide()
	}
	return &floatingRenderer{WidgetRenderer: widget.NewSimpleRenderer(container.NewWithoutLayout(img, circle, icon)), owner: i, img: img, circle: circle, icon: icon}
}

type floatingRenderer struct {
	fyne.WidgetRenderer
	owner  *indicatorWidget
	img    *canvas.Image
	circle *canvas.Circle
	icon   *canvas.Image
}

func (r *floatingRenderer) MinSize() fyne.Size    { return fyne.NewSize(184, 88) }
func (r *floatingRenderer) Layout(size fyne.Size) { r.img.Resize(size) }
func (r *floatingRenderer) Refresh() {
	r.img.Image = r.owner.image
	r.img.Refresh()
	if r.owner.hover {
		r.circle.Show()
		r.icon.Show()
	} else {
		r.circle.Hide()
		r.icon.Hide()
	}
}
func (i *indicatorWidget) MouseIn(*desktop.MouseEvent)    { i.hover = true; i.Refresh() }
func (i *indicatorWidget) MouseMoved(*desktop.MouseEvent) {}
func (i *indicatorWidget) MouseOut()                      { i.hover = false; i.Refresh() }
func (i *indicatorWidget) Tapped(e *fyne.PointEvent) {
	if i.hover && e.Position.X >= 10 && e.Position.X <= 28 && e.Position.Y >= 8 && e.Position.Y <= 26 {
		i.hide()
	}
}
func (i *indicatorWidget) Dragged(e *fyne.DragEvent) {
	platform.Drag(i.window, e.Dragged.DX, e.Dragged.DY)
}
func (i *indicatorWidget) DragEnd()                         {}
func (i *indicatorWidget) TappedSecondary(*fyne.PointEvent) { i.hide() }

type windowControl struct {
	widget.BaseWidget
	resource fyne.Resource
	action   func()
	hover    bool
}

func newWindowControl(resource fyne.Resource, action func()) *windowControl {
	b := &windowControl{resource: resource, action: action}
	b.ExtendBaseWidget(b)
	return b
}
func (b *windowControl) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	bg.CornerRadius = 4
	icon := canvas.NewImageFromResource(b.resource)
	icon.SetMinSize(fyne.NewSize(10, 10))
	return &controlRenderer{WidgetRenderer: widget.NewSimpleRenderer(container.NewWithoutLayout(bg, icon)), button: b, bg: bg, icon: icon}
}

type controlRenderer struct {
	fyne.WidgetRenderer
	button *windowControl
	bg     *canvas.Rectangle
	icon   *canvas.Image
}

func (r *controlRenderer) MinSize() fyne.Size { return fyne.NewSize(18, 18) }
func (r *controlRenderer) Layout(s fyne.Size) {
	r.bg.Resize(s)
	r.icon.Resize(fyne.NewSize(10, 10))
	r.icon.Move(fyne.NewPos((s.Width-10)/2, (s.Height-10)/2))
}
func (r *controlRenderer) Refresh() {
	name := theme.ColorNameButton
	if r.button.hover {
		name = theme.ColorNameHover
	}
	r.bg.FillColor = theme.Color(name)
	r.bg.Refresh()
	r.icon.Refresh()
}
func (b *windowControl) Tapped(*fyne.PointEvent)        { b.action() }
func (b *windowControl) MouseIn(*desktop.MouseEvent)    { b.hover = true; b.Refresh() }
func (b *windowControl) MouseMoved(*desktop.MouseEvent) {}
func (b *windowControl) MouseOut()                      { b.hover = false; b.Refresh() }
