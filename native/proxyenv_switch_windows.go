//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	appTitle   = "ProxyEnv Switch"
	appVersion = "1.2.2"

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
	SS_RIGHT         = 0x00000002
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
	WM_DPICHANGED    = 0x02E0
	EM_SETSEL        = 0x00B1

	COLOR_WINDOW = 5
	IDC_ARROW    = 32512

	MB_OK          = 0x00000000
	MB_ICONERROR   = 0x00000010
	MB_ICONINFO    = 0x00000040
	MB_ICONWARNING = 0x00000030

	HWND_BROADCAST   = 0xffff
	SMTO_ABORTIFHUNG = 0x0002
	SMTO_ERRORONEXIT = 0x0020

	environmentBroadcastTimeoutMS = 1000

	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010

	FW_NORMAL   = 400
	FW_SEMIBOLD = 600

	ID_PORT   = 1001
	ID_ADD    = 1002
	ID_REMOVE = 1003

	KEY_QUERY_VALUE      = 0x0001
	KEY_SET_VALUE        = 0x0002
	REG_SZ               = 1
	REG_EXPAND_SZ        = 2
	ERROR_SUCCESS        = 0
	ERROR_FILE_NOT_FOUND = 2
)

var proxyVariables = []string{"HTTP_PROXY", "HTTPS_PROXY"}

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

type savedValue struct {
	value  string
	exists bool
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

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
	procGetDpiForWindow        = user32.NewProc("GetDpiForWindow")
	procGetDpiForSystem        = user32.NewProc("GetDpiForSystem")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procSetWindowPos           = user32.NewProc("SetWindowPos")
	procGetSystemMetrics       = user32.NewProc("GetSystemMetrics")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procCreateFontW      = gdi32.NewProc("CreateFontW")
	procDeleteObject     = gdi32.NewProc("DeleteObject")

	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegCreateKeyExW  = advapi32.NewProc("RegCreateKeyExW")
	procRegSetValueExW   = advapi32.NewProc("RegSetValueExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procRegDeleteValueW  = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey      = advapi32.NewProc("RegCloseKey")

	hwndTitle, hwndVersion, hwndSubtitle, hwndAddressLabel, hwndPort syscall.Handle
	hwndAdd, hwndRemove, hwndGroup, hwndHTTP, hwndHTTPS              syscall.Handle
	hwndStatus, hwndNote                                             syscall.Handle
	fontNormal, fontTitle, fontSmall                                 syscall.Handle
	currentDPI                                                       int32 = 96
)

func utf16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func loword(v uintptr) uint16 { return uint16(v & 0xffff) }

func scale(v int32) int32 {
	return (v*currentDPI + 48) / 96
}

func createControl(className, text string, style uint32, parent syscall.Handle, id uintptr, hInstance syscall.Handle) syscall.Handle {
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(style),
		0, 0, 0, 0,
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
	procMessageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(unsafe.Pointer(utf16Ptr(appTitle))),
		flags,
	)
}

func move(hwnd syscall.Handle, x, y, w, h int32) {
	procMoveWindow.Call(uintptr(hwnd), uintptr(x), uintptr(y), uintptr(w), uintptr(h), 1)
}

func applyFont(hwnd, font syscall.Handle) {
	if hwnd != 0 && font != 0 {
		procSendMessageW.Call(uintptr(hwnd), WM_SETFONT, uintptr(font), 1)
	}
}

func createFont(height int32, weight int32) syscall.Handle {
	h, _, _ := procCreateFontW.Call(
		uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0,
		1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))),
	)
	return syscall.Handle(h)
}

func recreateFonts() {
	for _, font := range []syscall.Handle{fontNormal, fontTitle, fontSmall} {
		if font != 0 {
			procDeleteObject.Call(uintptr(font))
		}
	}
	fontNormal = createFont(-scale(16), FW_NORMAL)
	fontTitle = createFont(-scale(27), FW_SEMIBOLD)
	fontSmall = createFont(-scale(14), FW_NORMAL)

	for _, control := range []syscall.Handle{
		hwndSubtitle, hwndAddressLabel, hwndPort, hwndAdd, hwndRemove,
		hwndGroup, hwndHTTP, hwndHTTPS, hwndStatus,
	} {
		applyFont(control, fontNormal)
	}
	applyFont(hwndTitle, fontTitle)
	applyFont(hwndVersion, fontSmall)
	applyFont(hwndNote, fontSmall)
}

func registryError(code uintptr, operation string) error {
	return fmt.Errorf("%s: %w", operation, syscall.Errno(code))
}

func openEnvironmentKey(access uint32, create bool) (syscall.Handle, error) {
	var key syscall.Handle
	if create {
		var disposition uint32
		result, _, _ := procRegCreateKeyExW.Call(
			uintptr(syscall.HKEY_CURRENT_USER),
			uintptr(unsafe.Pointer(utf16Ptr("Environment"))),
			0,
			0,
			0,
			uintptr(access),
			0,
			uintptr(unsafe.Pointer(&key)),
			uintptr(unsafe.Pointer(&disposition)),
		)
		if result != ERROR_SUCCESS {
			return 0, registryError(result, "open HKEY_CURRENT_USER\\Environment")
		}
		return key, nil
	}

	result, _, _ := procRegOpenKeyExW.Call(
		uintptr(syscall.HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(utf16Ptr("Environment"))),
		0,
		uintptr(access),
		uintptr(unsafe.Pointer(&key)),
	)
	if result != ERROR_SUCCESS {
		return 0, registryError(result, "open HKEY_CURRENT_USER\\Environment")
	}
	return key, nil
}

func closeRegistryKey(key syscall.Handle) {
	if key != 0 {
		procRegCloseKey.Call(uintptr(key))
	}
}

func queryUserVariable(name string) (string, bool, error) {
	key, err := openEnvironmentKey(KEY_QUERY_VALUE, false)
	if err != nil {
		if errors.Is(err, syscall.Errno(ERROR_FILE_NOT_FOUND)) {
			return "", false, nil
		}
		return "", false, err
	}
	defer closeRegistryKey(key)

	var valueType uint32
	var size uint32
	result, _, _ := procRegQueryValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if result == ERROR_FILE_NOT_FOUND {
		return "", false, nil
	}
	if result != ERROR_SUCCESS {
		return "", false, registryError(result, "read "+name)
	}
	if valueType != REG_SZ && valueType != REG_EXPAND_SZ {
		return "", false, fmt.Errorf("read %s: unsupported registry value type %d", name, valueType)
	}
	if size == 0 {
		return "", true, nil
	}

	buffer := make([]uint16, (size+1)/2)
	result, _, _ = procRegQueryValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if result != ERROR_SUCCESS {
		return "", false, registryError(result, "read "+name)
	}
	return syscall.UTF16ToString(buffer), true, nil
}

func setUserVariable(name, value string) error {
	key, err := openEnvironmentKey(KEY_SET_VALUE, true)
	if err != nil {
		return err
	}
	defer closeRegistryKey(key)

	data, err := syscall.UTF16FromString(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	result, _, _ := procRegSetValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
		0,
		REG_SZ,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)*2),
	)
	if result != ERROR_SUCCESS {
		return registryError(result, "write "+name)
	}
	return nil
}

func deleteUserVariable(name string) (bool, error) {
	key, err := openEnvironmentKey(KEY_SET_VALUE, false)
	if err != nil {
		if errors.Is(err, syscall.Errno(ERROR_FILE_NOT_FOUND)) {
			return false, nil
		}
		return false, err
	}
	defer closeRegistryKey(key)

	result, _, _ := procRegDeleteValueW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(utf16Ptr(name))),
	)
	if result == ERROR_FILE_NOT_FOUND {
		return false, nil
	}
	if result != ERROR_SUCCESS {
		return false, registryError(result, "delete "+name)
	}
	return true, nil
}

func snapshotProxyVariables() (map[string]savedValue, error) {
	snapshot := make(map[string]savedValue, len(proxyVariables))
	for _, name := range proxyVariables {
		value, exists, err := queryUserVariable(name)
		if err != nil {
			return nil, err
		}
		snapshot[name] = savedValue{value: value, exists: exists}
	}
	return snapshot, nil
}

func restoreProxyVariables(snapshot map[string]savedValue) error {
	var failures []string
	for _, name := range proxyVariables {
		previous := snapshot[name]
		var err error
		if previous.exists {
			err = setUserVariable(name, previous.value)
		} else {
			_, err = deleteUserVariable(name)
		}
		if err != nil {
			failures = append(failures, name+": "+err.Error())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

func setProxyValues(proxyURL string) error {
	previous, err := snapshotProxyVariables()
	if err != nil {
		return err
	}
	for _, name := range proxyVariables {
		if err := setUserVariable(name, proxyURL); err != nil {
			if rollbackErr := restoreProxyVariables(previous); rollbackErr != nil {
				return fmt.Errorf("%v; rollback also failed: %v", err, rollbackErr)
			}
			return err
		}
	}
	for _, name := range proxyVariables {
		value, exists, err := queryUserVariable(name)
		if err != nil || !exists || value != proxyURL {
			verifyErr := err
			if verifyErr == nil {
				verifyErr = fmt.Errorf("Windows did not preserve the expected %s value", name)
			}
			if rollbackErr := restoreProxyVariables(previous); rollbackErr != nil {
				return fmt.Errorf("%v; rollback also failed: %v", verifyErr, rollbackErr)
			}
			return verifyErr
		}
	}
	return nil
}

func removeProxyValues() (bool, error) {
	previous, err := snapshotProxyVariables()
	if err != nil {
		return false, err
	}
	existed := false
	for _, saved := range previous {
		existed = existed || saved.exists
	}
	for _, name := range proxyVariables {
		if _, err := deleteUserVariable(name); err != nil {
			if rollbackErr := restoreProxyVariables(previous); rollbackErr != nil {
				return existed, fmt.Errorf("%v; rollback also failed: %v", err, rollbackErr)
			}
			return existed, err
		}
	}
	for _, name := range proxyVariables {
		_, exists, err := queryUserVariable(name)
		if err != nil || exists {
			verifyErr := err
			if verifyErr == nil {
				verifyErr = fmt.Errorf("Windows did not remove %s", name)
			}
			if rollbackErr := restoreProxyVariables(previous); rollbackErr != nil {
				return existed, fmt.Errorf("%v; rollback also failed: %v", verifyErr, rollbackErr)
			}
			return existed, verifyErr
		}
	}
	return existed, nil
}

func displayValue(name string) string {
	value, exists, err := queryUserVariable(name)
	if err != nil {
		return "(read error)"
	}
	if !exists {
		return "(not set)"
	}
	if value == "" {
		return "(empty value)"
	}
	return value
}

func refreshCurrentValues() {
	setText(hwndHTTP, "HTTP_PROXY:   "+displayValue("HTTP_PROXY"))
	setText(hwndHTTPS, "HTTPS_PROXY:  "+displayValue("HTTPS_PROXY"))
}

func addProxy(hwnd syscall.Handle) {
	raw := strings.TrimSpace(getText(hwndPort))
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		messageBox(hwnd, "Enter a numeric port from 1 to 65535.", MB_OK|MB_ICONWARNING)
		return
	}

	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := setProxyValues(proxyURL); err != nil {
		messageBox(hwnd, "Windows could not update both variables safely:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		refreshCurrentValues()
		return
	}

	_ = os.Setenv("HTTP_PROXY", proxyURL)
	_ = os.Setenv("HTTPS_PROXY", proxyURL)
	notifyEnvironmentChangeAsync()
	refreshCurrentValues()
	status := "Proxy variables were added successfully. Windows is being notified in the background. Reopen applications or terminals that were already running so they inherit the new values."
	setText(hwndStatus, status)
	messageBox(hwnd, "HTTP_PROXY and HTTPS_PROXY were set to:\n\n"+proxyURL, MB_OK|MB_ICONINFO)
}

func removeProxy(hwnd syscall.Handle) {
	existed, err := removeProxyValues()
	if err != nil {
		messageBox(hwnd, "Windows could not remove both variables safely:\n\n"+err.Error(), MB_OK|MB_ICONERROR)
		refreshCurrentValues()
		return
	}
	_ = os.Unsetenv("HTTP_PROXY")
	_ = os.Unsetenv("HTTPS_PROXY")
	notifyEnvironmentChangeAsync()
	refreshCurrentValues()

	if !existed {
		setText(hwndStatus, "No user-level HTTP_PROXY or HTTPS_PROXY values were present. Nothing needed to be removed.")
		messageBox(hwnd, "The proxy variables were already absent.", MB_OK|MB_ICONINFO)
		return
	}
	status := "Proxy variables were removed successfully. Windows is being notified in the background. Reopen applications or terminals that were already running to clear inherited values."
	setText(hwndStatus, status)
	messageBox(hwnd, "HTTP_PROXY and HTTPS_PROXY were removed from the current user's environment.", MB_OK|MB_ICONINFO)
}

func notifyEnvironmentChangeAsync() {
	go func() {
		_ = broadcastEnvironmentChange()
	}()
}

func broadcastEnvironmentChange() bool {
	var result uintptr
	ret, _, _ := procSendMessageTimeoutW.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(utf16Ptr("Environment"))),
		SMTO_ABORTIFHUNG|SMTO_ERRORONEXIT,
		environmentBroadcastTimeoutMS,
		uintptr(unsafe.Pointer(&result)),
	)
	return ret != 0
}

func layout(hwnd syscall.Handle) {
	var r rect
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	width := r.right - r.left
	height := r.bottom - r.top
	margin := scale(28)
	contentW := width - margin*2
	if contentW < scale(360) {
		contentW = scale(360)
	}

	move(hwndTitle, margin, scale(20), contentW-scale(80), scale(38))
	move(hwndVersion, width-margin-scale(72), scale(25), scale(72), scale(25))
	move(hwndSubtitle, margin, scale(64), contentW, scale(56))
	move(hwndAddressLabel, margin, scale(128), scale(255), scale(30))
	move(hwndPort, margin+scale(255), scale(124), scale(112), scale(32))

	gap := scale(14)
	buttonW := (contentW - gap) / 2
	move(hwndAdd, margin, scale(176), buttonW, scale(42))
	move(hwndRemove, margin+buttonW+gap, scale(176), buttonW, scale(42))

	groupY := scale(238)
	groupH := scale(112)
	move(hwndGroup, margin, groupY, contentW, groupH)
	move(hwndHTTP, margin+scale(18), groupY+scale(32), contentW-scale(36), scale(28))
	move(hwndHTTPS, margin+scale(18), groupY+scale(64), contentW-scale(36), scale(28))

	noteH := scale(58)
	statusY := groupY + groupH + scale(18)
	statusH := height - statusY - noteH - scale(18)
	if statusH < scale(70) {
		statusH = scale(70)
	}
	move(hwndStatus, margin, statusY, contentW, statusH)
	move(hwndNote, margin, height-noteH-scale(8), contentW, noteH)
}

func getDPIForWindow(hwnd syscall.Handle) int32 {
	if err := procGetDpiForWindow.Find(); err == nil {
		dpi, _, _ := procGetDpiForWindow.Call(uintptr(hwnd))
		if dpi >= 96 {
			return int32(dpi)
		}
	}
	return currentDPI
}

func getSystemDPI() int32 {
	if err := procGetDpiForSystem.Find(); err == nil {
		dpi, _, _ := procGetDpiForSystem.Call()
		if dpi >= 96 {
			return int32(dpi)
		}
	}
	return 96
}

func wndProc(hwnd syscall.Handle, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_CREATE:
		hInstanceRaw, _, _ := procGetModuleHandleW.Call(0)
		hInstance := syscall.Handle(hInstanceRaw)
		currentDPI = getDPIForWindow(hwnd)

		hwndTitle = createControl("STATIC", "Windows Proxy Environment Variables", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndVersion = createControl("STATIC", "v"+appVersion, WS_CHILD|WS_VISIBLE|SS_RIGHT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndSubtitle = createControl("STATIC", "Add, update, or remove persistent user-level HTTP_PROXY and HTTPS_PROXY values. Administrator rights are not required.", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndAddressLabel = createControl("STATIC", "Proxy address:  http://127.0.0.1:", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndPort = createControl("EDIT", "60505", WS_CHILD|WS_VISIBLE|WS_TABSTOP|WS_BORDER|ES_LEFT|ES_NUMBER, hwnd, ID_PORT, hInstance)
		hwndAdd = createControl("BUTTON", "Add / Update Proxy", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, hwnd, ID_ADD, hInstance)
		hwndRemove = createControl("BUTTON", "Remove Proxy", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, hwnd, ID_REMOVE, hInstance)
		hwndGroup = createControl("BUTTON", "Current user values", WS_CHILD|WS_VISIBLE|BS_GROUPBOX, hwnd, 0, hInstance)
		hwndHTTP = createControl("STATIC", "HTTP_PROXY:   (not set)", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndHTTPS = createControl("STATIC", "HTTPS_PROXY:  (not set)", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndStatus = createControl("STATIC", "Enter a port, then choose an action.", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)
		hwndNote = createControl("STATIC", "This changes environment variables only; it does not change the Windows system proxy. Reopen existing terminals and applications after a change.", WS_CHILD|WS_VISIBLE|SS_LEFT|SS_NOPREFIX, hwnd, 0, hInstance)

		recreateFonts()
		layout(hwnd)
		refreshCurrentValues()
		procSetFocus.Call(uintptr(hwndPort))
		procSendMessageW.Call(uintptr(hwndPort), EM_SETSEL, 0, ^uintptr(0))
		return 0

	case WM_SIZE:
		layout(hwnd)
		return 0

	case WM_DPICHANGED:
		newDPI := int32(loword(wParam))
		if newDPI >= 96 {
			currentDPI = newDPI
		}
		suggested := (*rect)(unsafe.Pointer(lParam))
		procSetWindowPos.Call(
			uintptr(hwnd),
			0,
			uintptr(suggested.left),
			uintptr(suggested.top),
			uintptr(suggested.right-suggested.left),
			uintptr(suggested.bottom-suggested.top),
			SWP_NOZORDER|SWP_NOACTIVATE,
		)
		recreateFonts()
		layout(hwnd)
		return 0

	case WM_GETMINMAXINFO:
		mmi := (*minMaxInfo)(unsafe.Pointer(lParam))
		mmi.minTrackSize.x = scale(680)
		mmi.minTrackSize.y = scale(500)
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
		for _, font := range []syscall.Handle{fontNormal, fontTitle, fontSmall} {
			if font != 0 {
				procDeleteObject.Call(uintptr(font))
			}
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(message), wParam, lParam)
	return ret
}

func enableDPIAwareness() {
	// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 is the pointer value -4.
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
	procSetWindowPos.Call(uintptr(hwnd), 0, uintptr(x), uintptr(y), uintptr(w), uintptr(h), SWP_NOZORDER)
}

func main() {
	enableDPIAwareness()
	currentDPI = getSystemDPI()

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
		uintptr(unsafe.Pointer(utf16Ptr(appTitle+" "+appVersion))),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE|WS_CLIPCHILDREN,
		CW_USEDEFAULT, CW_USEDEFAULT,
		uintptr(scale(800)), uintptr(scale(570)),
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
