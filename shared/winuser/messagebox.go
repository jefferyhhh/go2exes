package winuser

import (
	"syscall"
	"unsafe"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

const (
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
	MB_ICONERROR       = 0x00000010
	MB_ICONWARNING     = 0x00000030
	MB_TOPMOST         = 0x00040000
)

// ShowMessage 显示一个消息框
// title: 标题  msg: 内容  isError: 是否为错误图标
func ShowMessage(title, msg string, isError bool) {
	text, _ := syscall.UTF16PtrFromString(msg)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	flags := uintptr(MB_OK | MB_TOPMOST)
	if isError {
		flags |= MB_ICONERROR
	} else {
		flags |= MB_ICONINFORMATION
	}

	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(text)),
		uintptr(unsafe.Pointer(titlePtr)),
		flags,
	)
}

// ShowInfo 显示信息提示框
func ShowInfo(title, msg string) {
	ShowMessage(title, msg, false)
}

// ShowError 显示错误提示框
func ShowError(title, msg string) {
	ShowMessage(title, msg, true)
}

// ShowWarning 显示警告提示框
func ShowWarning(title, msg string) {
	text, _ := syscall.UTF16PtrFromString(msg)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	flags := uintptr(MB_OK | MB_TOPMOST | MB_ICONWARNING)
	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(text)),
		uintptr(unsafe.Pointer(titlePtr)),
		flags,
	)
}
