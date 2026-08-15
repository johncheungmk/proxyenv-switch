//go:build windows

package main

import "runtime"

// Win32 window ownership and message queues are thread-affine. Go goroutines may
// otherwise move between OS threads, which can leave the window on one Windows
// thread while GetMessageW runs on another. Lock the main goroutine before any
// window is created so creation, focus/input handling, and the message loop all
// remain on the same OS thread for the lifetime of the application.
func init() {
	runtime.LockOSThread()
}
