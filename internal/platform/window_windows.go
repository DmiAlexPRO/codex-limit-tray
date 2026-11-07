package platform

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"sync"
	"syscall"
	"unsafe"
)

var regionWindows sync.Map
var regionWindowProc = syscall.NewCallback(func(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	entry, ok := regionWindows.Load(hwnd)
	if !ok {
		return 0
	}
	state := entry.(regionState)
	result, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("CallWindowProcW").Call(state.previous, hwnd, uintptr(msg), wparam, lparam)
	if msg == 0x0005 || msg == 0x02E0 {
		applyRegion(hwnd, state.radius)
	}
	if msg == 0x0082 {
		regionWindows.Delete(hwnd)
	}
	return result
})

type regionState struct {
	previous uintptr
	radius   int
}

// Create the main window without native decorations from the outset so GLFW
// uses the correct client-area dimensions for all subsequent resize requests.
func NewMainWindow(app fyne.App, title string) fyne.Window {
	if d, ok := app.Driver().(desktop.Driver); ok {
		w := d.CreateSplashWindow()
		w.SetTitle(title)
		return w
	}
	return app.NewWindow(title)
}

func PrepareWindow(w fyne.Window) {
	configure(w, false)
}

func PrepareMainWindow(w fyne.Window) { configure(w, true) }

func configure(w fyne.Window, main bool) {
	native, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	native.RunNative(func(context interface{}) {
		win, ok := context.(driver.WindowsWindowContext)
		if !ok {
			return
		}
		user := syscall.NewLazyDLL("user32.dll")
		style, _, _ := user.NewProc("GetWindowLongPtrW").Call(win.HWND, ^uintptr(19))
		style = (style | 0x80) &^ 0x40000
		if main {
			style = (style | 0x40000) &^ 0x80
		}
		_, _, _ = user.NewProc("SetWindowLongPtrW").Call(win.HWND, ^uintptr(19), style)
		_, _, _ = user.NewProc("SetWindowPos").Call(win.HWND, 0, 0, 0, 0, 0, 0x37)
		if main {
			user.NewProc("ShowWindow").Call(win.HWND, 9)
			user.NewProc("SetForegroundWindow").Call(win.HWND)
		}
	})
	radius := 32
	if main {
		radius = 12
	}
	roundWindow(w, radius)
}

func roundWindow(w fyne.Window, radius int) {
	withHandle(w, func(h uintptr) {
		if _, exists := regionWindows.Load(h); !exists {
			user := syscall.NewLazyDLL("user32.dll")
			previous, _, _ := user.NewProc("SetWindowLongPtrW").Call(h, ^uintptr(3), regionWindowProc)
			if previous != 0 {
				regionWindows.Store(h, regionState{previous: previous, radius: radius})
			}
		}
		applyRegion(h, radius)
	})
}

func applyRegion(h uintptr, radius int) {
	user := syscall.NewLazyDLL("user32.dll")
	gdi := syscall.NewLazyDLL("gdi32.dll")
	var rect struct{ Left, Top, Right, Bottom int32 }
	ok, _, _ := user.NewProc("GetWindowRect").Call(h, uintptr(unsafe.Pointer(&rect)))
	if ok == 0 {
		return
	}
	dpi, _, _ := user.NewProc("GetDpiForWindow").Call(h)
	if dpi == 0 {
		dpi = 96
	}
	diameter := uintptr(radius*2) * dpi / 96
	region, _, _ := gdi.NewProc("CreateRoundRectRgn").Call(0, 0, uintptr(rect.Right-rect.Left+1), uintptr(rect.Bottom-rect.Top+1), diameter, diameter)
	if region == 0 {
		return
	}
	applied, _, _ := user.NewProc("SetWindowRgn").Call(h, region, 1)
	if applied == 0 {
		gdi.NewProc("DeleteObject").Call(region)
	}
}

func Sound() { syscall.NewLazyDLL("user32.dll").NewProc("MessageBeep").Call(0x40) }

func Minimize(w fyne.Window) {
	withHandle(w, func(h uintptr) { syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow").Call(h, 6) })
}
func Drag(w fyne.Window, dx, dy float32) {
	withHandle(w, func(h uintptr) {
		user := syscall.NewLazyDLL("user32.dll")
		user.NewProc("ReleaseCapture").Call()
		user.NewProc("SendMessageW").Call(h, 0xA1, 2, 0)
	})
}
func withHandle(w fyne.Window, fn func(uintptr)) {
	if native, ok := w.(driver.NativeWindow); ok {
		native.RunNative(func(c interface{}) {
			if win, ok := c.(driver.WindowsWindowContext); ok && win.HWND != 0 {
				fn(win.HWND)
			}
		})
	}
}
