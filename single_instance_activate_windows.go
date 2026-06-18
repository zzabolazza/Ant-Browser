//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	swRestore = 9
)

var (
	user32Activate               = windows.NewLazySystemDLL("user32.dll")
	procAllowSetForegroundWindow = user32Activate.NewProc("AllowSetForegroundWindow")
	procEnumWindows              = user32Activate.NewProc("EnumWindows")
	procGetWindowThreadProcessID = user32Activate.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32Activate.NewProc("IsWindowVisible")
	procShowWindow               = user32Activate.NewProc("ShowWindow")
	procSetForegroundWindow      = user32Activate.NewProc("SetForegroundWindow")
)

func activateExistingSingleInstanceWindow(pid int) {
	if pid <= 0 {
		return
	}
	procAllowSetForegroundWindow.Call(uintptr(pid))
	if hwnd := findTopLevelWindowByPID(uint32(pid)); hwnd != 0 {
		procShowWindow.Call(hwnd, swRestore)
		procSetForegroundWindow.Call(hwnd)
	}
}

func findTopLevelWindowByPID(pid uint32) uintptr {
	var matched uintptr
	callback := windows.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible == 0 {
			return 1
		}

		var windowPID uint32
		procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&windowPID)))
		if windowPID == pid {
			matched = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(callback, 0)
	return matched
}
