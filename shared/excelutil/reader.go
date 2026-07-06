package excelutil

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ColumnData 列数据结构
type ColumnData struct {
	Headers []string
	Rows    [][]string
}

// ReadColumns 从 Excel 读取指定列
// cols 格式如 "A:B"
func ReadColumns(filePath, sheetName, cols string) (*ColumnData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 失败: %w", err)
	}
	defer f.Close()

	if sheetName == "" {
		sheets := f.GetSheetList()
		if len(sheets) > 0 {
			sheetName = sheets[0]
		}
	}

	colParts := strings.Split(cols, ":")
	if len(colParts) != 2 {
		return nil, fmt.Errorf("列范围格式应为 A:B")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	startIdx := colNameToIndex(strings.TrimSpace(strings.ToUpper(colParts[0])))
	endIdx := colNameToIndex(strings.TrimSpace(strings.ToUpper(colParts[1])))

	data := &ColumnData{
		Headers: make([]string, 0),
		Rows:    make([][]string, 0),
	}

	if len(rows) > 0 {
		for i := startIdx; i <= endIdx && i < len(rows[0]); i++ {
			data.Headers = append(data.Headers, rows[0][i])
		}
	}

	for i := 1; i < len(rows); i++ {
		rowData := make([]string, 0)
		for j := startIdx; j <= endIdx && j < len(rows[i]); j++ {
			rowData = append(rowData, rows[i][j])
		}
		data.Rows = append(data.Rows, rowData)
	}

	return data, nil
}

func colNameToIndex(col string) int {
	result := 0
	for _, c := range col {
		result = result*26 + int(c-'A') + 1
	}
	return result - 1
}
