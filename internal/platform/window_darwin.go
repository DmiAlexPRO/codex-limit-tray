package platform

/*
#cgo LDFLAGS: -framework Cocoa
void ctlAccessory(void);
*/
import "C"
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

func PrepareWindow(w fyne.Window) {
	if native, ok := w.(driver.NativeWindow); ok {
		native.RunNative(func(context interface{}) { C.ctlAccessory() })
	}
}
