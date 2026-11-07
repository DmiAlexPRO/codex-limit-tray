//go:build !windows && !darwin

package platform

import "fyne.io/fyne/v2"

func PrepareWindow(w fyne.Window) {}
