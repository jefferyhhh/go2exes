package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/xuri/excelize/v2"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

const (
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
	MB_ICONERROR       = 0x00000010
	MB_TOPMOST         = 0x00040000
)

func main() {
	// 获取 exe 所在目录
	exePath, err := os.Executable()
	if err != nil {
		showMsg(fmt.Sprintf("获取程序路径失败: %v", err), true)
		return
	}
	dir := filepath.Dir(exePath)

	// 如果有命令行参数，使用参数指定的目录
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// 检查目录是否存在
	info, err := os.Stat(dir)
	if err != nil {
		showMsg(fmt.Sprintf("目录不存在: %s", dir), true)
		return
	}
	if !info.IsDir() {
		showMsg(fmt.Sprintf("路径不是目录: %s", dir), true)
		return
	}

	// 遍历目录下的 Excel 文件
	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".xlsx" || ext == ".xls" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		showMsg(fmt.Sprintf("遍历目录失败: %v", err), true)
		return
	}

	if len(files) == 0 {
		showMsg("未找到 Excel 文件", false)
		return
	}

	// 处理每个文件
	var success, failed int
	var errors []string
	for _, file := range files {
		if err := processFile(file); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(file), err))
		} else {
			success++
		}
	}

	// 显示结果
	msg := fmt.Sprintf("处理完成！\n\n目录: %s\n\n找到文件: %d\n成功: %d\n失败: %d", dir, len(files), success, failed)
	if len(errors) > 0 {
		msg += "\n\n失败详情:\n" + strings.Join(errors, "\n")
	}
	showMsg(msg, false)
}

func showMsg(msg string, isError bool) {
	title, _ := syscall.UTF16PtrFromString("空行去除工具")
	text, _ := syscall.UTF16PtrFromString(msg)

	flags := uintptr(MB_OK | MB_TOPMOST)
	if isError {
		flags |= MB_ICONERROR
	} else {
		flags |= MB_ICONINFORMATION
	}

	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(text)),
		uintptr(unsafe.Pointer(title)),
		flags,
	)
}

func processFile(filePath string) error {
	// 备份原文件
	backupPath := filePath + ".bak"
	if err := copyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 处理每个 sheet
	for _, sheetName := range f.GetSheetList() {
		if err := removeEmptyRows(f, sheetName); err != nil {
			return fmt.Errorf("处理 sheet %s 失败: %w", sheetName, err)
		}
	}

	// 保存文件
	if err := f.Save(); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}

func removeEmptyRows(f *excelize.File, sheetName string) error {
	// 获取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	// 过滤空行
	var nonEmptyRows [][]string
	for _, row := range rows {
		if !isRowEmpty(row) {
			nonEmptyRows = append(nonEmptyRows, row)
		}
	}

	// 获取 sheet 的最大行数和列数
	maxRow := len(rows)
	maxCol := 0
	for _, row := range rows {
		if len(row) > maxCol {
			maxCol = len(row)
		}
	}

	// 清空所有单元格
	for i := 1; i <= maxRow; i++ {
		for j := 1; j <= maxCol; j++ {
			cellName, _ := excelize.CoordinatesToCellName(j, i)
			f.SetCellValue(sheetName, cellName, "")
		}
	}

	// 写入非空行
	for i, row := range nonEmptyRows {
		for j, cell := range row {
			cellName, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue(sheetName, cellName, cell)
		}
	}

	return nil
}

func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
