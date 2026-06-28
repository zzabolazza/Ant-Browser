//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	swShow    = 5
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
	procSetWindowPos             = user32Activate.NewProc("SetWindowPos")
)

const (
	hwndTopmost   = ^uintptr(0)
	hwndNotopmost = ^uintptr(1)
	swpNoMove     = 0x0002
	swpNoSize     = 0x0001
	swpShowWindow = 0x0040
)

func grantExistingSingleInstanceForeground(pid int) {
	if pid > 0 {
		procAllowSetForegroundWindow.Call(uintptr(pid))
	}
}

func activateExistingSingleInstanceWindow(pid int) {
	if pid <= 0 {
		return
	}
	grantExistingSingleInstanceForeground(pid)
	if hwnd := findTopLevelWindowByPID(uint32(pid)); hwnd != 0 {
		procShowWindow.Call(hwnd, swShow)
		procShowWindow.Call(hwnd, swRestore)
		if ok, _, _ := procSetForegroundWindow.Call(hwnd); ok == 0 {
			procSetWindowPos.Call(hwnd, hwndTopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpShowWindow)
			procSetWindowPos.Call(hwnd, hwndNotopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpShowWindow)
			procSetForegroundWindow.Call(hwnd)
		}
	}
}

func findTopLevelWindowByPID(pid uint32) uintptr {
	var matched uintptr
	callback := windows.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		var windowPID uint32
		procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&windowPID)))
		if windowPID == pid {
			if visible, _, _ := procIsWindowVisible.Call(hwnd); visible == 0 && matched != 0 {
				return 1
			}
			matched = hwnd
			if visible, _, _ := procIsWindowVisible.Call(hwnd); visible != 0 {
				return 0
			}
			return 1
		}
		return 1
	})
	procEnumWindows.Call(callback, 0)
	return matched
}
