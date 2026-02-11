package handler

import "fmt"

// cellName 将列号(1-based)和行号转换为Excel单元格名称，如 cellName(1, 2) => "A2"
func cellName(col, row int) string {
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}
	return fmt.Sprintf("%s%d", colName, row)
}
