//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	appTitle = "ProxyEnv Switch"

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_TABSTOP          = 0x00010000
	WS_BORDER           = 0x00800000
	WS_CLIPCHILDREN     = 0x02000000

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	BS_GROUPBOX      = 0x00000007
	ES_LEFT          = 0x0000
	ES_NUMBER        = 0x2000
	SS_LEFT          = 0x00000000
	SS_NOPREFIX      = 0x00000080

	CW_USEDEFAULT = ^uintptr(0x7fffffff)
	SW_SHOW       = 5

	WM_CREATE        = 0x0001
	WM_DESTROY       = 0x0002
	WM_SIZE          = 0x0005
	WM_COMMAND       = 0x0111
	WM_SETTINGCHANGE = 0x001A
	WM_GETMINMAXINFO = 0x0024
	WM_SETFONT       = 0x0030
	EM_SETSEL        = 0x00B1

	COLOR_WINDOW = 5
	IDC_ARROW    = 32512

	MB_OK          = 0x00000000
	MB_ICONERROR   = 0x00000010
	MB_ICONINFO    = 0x00000040
	MB_ICONWARNING = 0x00000030

	HWND_BROADCAST   = 0xffff
	SMTO_ABORTIFHUNG = 0x0002

	FW_NORMAL   = 400
	FW_SEMIBOLD = 600

	ID_PORT   = 1001
	ID_ADD    = 1002
	ID_REMOVE = 1003
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       syscall.Handle
}

type point struct{ x, y int32 }
type msg struct {
	hwnd           syscall.Handle
	message        uint32
	wParam, lParam uintptr
	time           uint32
	pt             point
}

type minMaxInfo struct {
	reserved     point
	maxSize      point
	maxPosition  point
	minTrackSize point
	maxTrackSize point
}

type rect struct{ left, top, right, bottom int32 }

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procShowWindow             = user32.NewProc("ShowWindow")
	procUpdateWindow           = user32.NewProc("UpdateWindow")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procLoadCursorW            = user32.NewProc("LoadCursorW")
	procGetWindowTextW         = user32.NewProc("GetWindowTextW")
	procSetWindowTextW         = user32.NewProc("SetWindowTextW")
	procMessageBoxW            = user32.NewProc("MessageBoxW")
	procSendMessageTimeoutW    = user32.NewProc("SendMessageTimeoutW")
	procSetFocus               = user32.NewProc("SetFocus")
	procSendMessageW           = user32.NewProc("SendMessageW")
	procMoveWindow             = user32.NewProc("MoveWindow")
	procGetClientRect          = user32.NewProc("GetClientRect")
	procSetProcessDPIAware     = user32.NewProc("SetProcessDPIAware")
	procSetProcessDpiAwareness = user32.NewProc("SetProcessDpiAwarenessContext")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procSetWindowPos           = user32.NewProc("SetWindowPos")
	procGetSystemMetrics       = user32.NewProc("GetSystemMetrics")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW      = gdi32.NewProc("CreateFontW")
	procDeleteObject     = gdi32.NewProc("DeleteObject")

	hwndTitle, hwndSubtitle, hwndAddressLabel, hwndPort syscall.Handle
	hwndAdd, hwndRemove, hwndGroup, hwndHTTP, hwndHTTPS syscall.Handle
	hwndStatus, hwndNote                                syscall.Handle
	fontNormal, fontTitle                               syscall.Handle
)

func utf16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func loword(v uintptr) uint16 { return uint16(v & 0xffff) }

func createControl(className, text string, style uint32, x, y, width, height int32, parent syscall.Handle, id uintptr, hInstance syscall.Handle) syscall.Handle {
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		uintptr(parent), id, uintptr(hInstance), 0,
	)
	return syscall.Handle(hwnd)
}

func setText(hwnd syscall.Handle, text string) {
	procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getText(hwnd syscall.Handle) string {
	buf := make([]uint16, 128)
	n, _, _ := procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

func messageBox(hwnd syscall.Handle, text string, flags uintptr) {
	procMessageBoxW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(appTitle))), flags)
}

func move(hwnd syscall.Handle, x, y, w, h int32) {
	procMoveWindow.Call(uintptr(hwnd), uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
}

func applyFont(hwnd, font syscall.Handle) {
	procSendMessageW.Call(uintptr(hwnd), WM_SETFONT, uintptr(font), 1)
}

func createFont(height int32, weight int32) syscall.Handle {
	h, _, _ := procCreateFontW.Call(
		uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0,
		1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))),
	)
	return syscall.Handle(h)
}

func runReg(args ...string) ([]byte, error) {
	cmd := exec.Command("reg.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%v\n%s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func queryUserVariable(name string) string {
	out, err := runReg("query", `HKCU\Environment`, "/v", name)
	if err != nil {
		return "(not set)"
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToUpper(line), strings.ToUpper(name)) {
			continue
		}
		upper := strings.ToUpper(line)
		idx := strings.Index(upper, "REG_SZ")
		if idx >= 0 {
			value := strings.TrimSpace(line[idx+len("REG_SZ"):])
			if value != "" {
				return value
			}
		}
	}
	return "(not set)"
}

func refreshCurrentValues() {
	setText(hwndHTTP, "HTTP_PROXY:   "+queryUserVariable("HTTP_PROXY"))
	setText(hwndHTTPS, "HTTPS_PROXY:  "+queryUserVariable("HTTPS_PROXY"))
}

func addProxy(hwnd syscall.Handle) {
	raw := strings.TrimSpace(getText(hwndPort))
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		messageBox(hwnd, "Enter a numeric port from 1 to 65535.", MB_OK|MB_ICONWARNING)
		return
	}

	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if _, err := runReg("add", `HKCU\Environment`, "/v", "HTTP_PROXY", "/t", "REG_SZ", "/d", proxyURL, "/f"); err != nil {
		messageBox(hwnd, "Could not set HTTP_PROXY:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	if _, err := runReg("add", `HKCU\Environment`, "/v", "HTTPS_PROXY", "/t", "REG_SZ", "/d", proxyURL, "/f"); err != nil {
		messageBox(hwnd, "Could not set HTTPS_PROXY:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		return
	}

	_ = os.Setenv("HTTP_PROXY", proxyURL)
	_ = os.Setenv("HTTPS_PROXY", proxyURL)
	broadcastEnvironmentChange()
	refreshCurrentValues()
	setText(hwndStatus, "Proxy variables were added successfully. Reopen applications or terminals that were already running so they inherit the new values.")
	messageBox(hwnd, "HTTP_PROXY and HTTPS_PROXY were set to:\n\n"+proxyURL, MB_OK|MB_ICONINFO)
}

func removeProxy(hwnd syscall.Handle) {
	_, httpErr := runReg("delete", `HKCU\Environment`, "/v", "HTTP_PROXY", "/f")
	_, httpsErr := runReg("delete", `HKCU\Environment`, "/v", "HTTPS_PROXY", "/f")
	_ = os.Unsetenv("HTTP_PROXY")
	_ = os.Unsetenv("HTTPS_PROXY")
	broadcastEnvironmentChange()
	refreshCurrentValues()

	if httpErr != nil && httpsErr != nil {
		setText(hwndStatus, "No user-level HTTP_PROXY or HTTPS_PROXY values were present. Nothing needed to be removed.")
		messageBox(hwnd, "The proxy variables were already absent.", MB_OK|MB_ICONINFO)
		return
	}
	setText(hwndStatus, "Proxy variables were removed successfully. Reopen applications or terminals that were already running to clear inherited values.")
	messageBox(hwnd, "HTTP_PROXY and HTTPS_PROXY were removed from the current user's environment.", MB_OK|MB_ICONINFO)
}

func broadcastEnvironmentChange() {
	var result uintptr
	procSendMessageTimeoutW.Call(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(utf16Ptr("Environment"))), SMTO_ABORTIFHUNG, 5000, uintptr(unsafe.Pointer(&result)))
}

func layout(hwnd syscall.Handle) {
	var r rect
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	width := r.right - r.left
	height := r.bottom - r.top
	margin := int32(24)
	contentW := width - margin*2
	if contentW < 300 {
		contentW = 300
	}

	move(hwndTitle, margin, 18, contentW, 34)
	move(hwndSubtitle, margin, 56, contentW, 52)
	move(hwndAddressLabel, margin, 118, 250, 28)
	move(hwndPort, margin+250, 114, 110, 30)
	move(hwndAdd, margin, 162, 190, 38)
	move(hwndRemove, margin+204, 162, 170, 38)

	groupY := int32(218)
	groupH := int32(104)
	move(hwndGroup, margin, groupY, contentW, groupH)
	move(hwndHTTP, margin+16, groupY+30, contentW-32, 26)
	move(hwndHTTPS, margin+16, groupY+58, contentW-32, 26)

	noteH := int32(46)
	statusY := groupY + groupH + 16
	statusH := height - statusY - noteH - 22
	if statusH < 58 {
		statusH = 58
	}
	move(hwndStatus, margin, statusY, contentW, statusH)
	move(hwndNote, margin, height-noteH-10, contentW, noteH)
}

func wndProc(hwnd syscall.Handle, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CREATE:
		hInstanceRaw, _, _ := procGetModuleHandleW.Call(0)
		hInstance := syscall.Handle(hInstanceRaw)

		fontNormal = createFont(-16, FW_NORMAL)
		fontTitle = createFont(-24, FW_SEMIBOLD)

		hwndTitle = createControl("STATIC", "Windows Proxy Environment Variables", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndSubtitle = createControl("STATIC", "Add, update, or remove persistent user-level HTTP_PROXY and HTTPS_PROXY values. Administrator rights are not required.", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndAddressLabel = createControl("STATIC", "Proxy address:  http://127.0.0.1:", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndPort = createControl("EDIT", "60505", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_LEFT|ES_NUMBER, 0, 0, 0, 0, hwnd, ID_PORT, hInstance)
		hwndAdd = createControl("BUTTON", "Add / Update Proxy", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, 0, 0, 0, 0, hwnd, ID_ADD, hInstance)
		hwndRemove = createControl("BUTTON", "Remove Proxy", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 0, 0, hwnd, ID_REMOVE, hInstance)
		hwndGroup = createControl("BUTTON", "Current user values", WS_CHILD|WS_VISIBLE|BS_GROUPBOX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndHTTP = createControl("STATIC", "HTTP_PROXY:   (not set)", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndHTTPS = createControl("STATIC", "HTTPS_PROXY:  (not set)", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndStatus = createControl("STATIC", "Enter a port, then choose an action.", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)
		hwndNote = createControl("STATIC", "Changes apply to the current Windows user only. Reopen existing terminals and applications after making a change.", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, 0, 0, 0, 0, hwnd, 0, hInstance)

		controls := []syscall.Handle{hwndSubtitle, hwndAddressLabel, hwndPort, hwndAdd, hwndRemove, hwndGroup, hwndHTTP, hwndHTTPS, hwndStatus, hwndNote}
		for _, c := range controls {
			applyFont(c, fontNormal)
		}
		applyFont(hwndTitle, fontTitle)

		layout(hwnd)
		refreshCurrentValues()
		procSetFocus.Call(uintptr(hwndPort))
		procSendMessageW.Call(uintptr(hwndPort), EM_SETSEL, 0, ^uintptr(0))
		return 0

	case WM_SIZE:
		layout(hwnd)
		return 0

	case WM_GETMINMAXINFO:
		mmi := (*minMaxInfo)(unsafe.Pointer(lParam))
		mmi.minTrackSize.x = 650
		mmi.minTrackSize.y = 475
		return 0

	case WM_COMMAND:
		switch loword(wParam) {
		case ID_ADD:
			addProxy(hwnd)
		case ID_REMOVE:
			removeProxy(hwnd)
		}
		return 0

	case WM_DESTROY:
		if fontNormal != 0 {
			procDeleteObject.Call(uintptr(fontNormal))
		}
		if fontTitle != 0 {
			procDeleteObject.Call(uintptr(fontTitle))
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return ret
}

func enableDPIAwareness() {
	// PER_MONITOR_AWARE_V2 = -4. The bit pattern below is uintptr(-4).
	if err := procSetProcessDpiAwareness.Find(); err == nil {
		result, _, _ := procSetProcessDpiAwareness.Call(^uintptr(3))
		if result != 0 {
			return
		}
	}
	procSetProcessDPIAware.Call()
}

func centerWindow(hwnd syscall.Handle) {
	var r rect
	procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	w := r.right - r.left
	h := r.bottom - r.top
	screenW, _, _ := procGetSystemMetrics.Call(0)
	screenH, _, _ := procGetSystemMetrics.Call(1)
	x := (int32(screenW) - w) / 2
	y := (int32(screenH) - h) / 2
	procSetWindowPos.Call(uintptr(hwnd), 0, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0x0004)
}

func main() {
	enableDPIAwareness()

	hInstanceRaw, _, _ := procGetModuleHandleW.Call(0)
	hInstance := syscall.Handle(hInstanceRaw)
	className := utf16Ptr("ProxyEnvSwitchWindow")
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		style:         CS_HREDRAW | CS_VREDRAW,
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInstance,
		hCursor:       syscall.Handle(cursor),
		hbrBackground: syscall.Handle(COLOR_WINDOW + 1),
		lpszClassName: className,
	}

	atom, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		messageBox(0, "Could not register the application window:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		return
	}

	hwndRaw, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(appTitle))),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE|WS_CLIPCHILDREN,
		CW_USEDEFAULT, CW_USEDEFAULT,
		760, 540,
		0, 0, uintptr(hInstance), 0,
	)
	if hwndRaw == 0 {
		messageBox(0, "Could not create the application window:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		return
	}

	hwnd := syscall.Handle(hwndRaw)
	centerWindow(hwnd)
	procShowWindow.Call(uintptr(hwnd), SW_SHOW)
	procUpdateWindow.Call(uintptr(hwnd))

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
