//go:build !windows

package platform

import "fyne.io/fyne/v2"

func NewMainWindow(app fyne.App, title string) fyne.Window { return app.NewWindow(title) }

func PrepareMainWindow(w fyne.Window)    {}
func Minimize(w fyne.Window)             { w.Hide() }
func Drag(w fyne.Window, dx, dy float32) {}
func Sound()                             {}
