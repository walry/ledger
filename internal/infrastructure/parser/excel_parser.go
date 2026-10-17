package parser

import (
	"bytes"
	"context"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/walry/ledger/internal/domain/report"
	reportservice "github.com/walry/ledger/internal/domain/report/service"
)

// ExcelParser 实现 reportservice.Parser，基于 github.com/xuri/excelize/v2。
type ExcelParser struct{}

// NewExcelParser 创建 Excel 解析器。
func NewExcelParser() *ExcelParser { return &ExcelParser{} }

// Supports 判断是否支持该格式。
func (p *ExcelParser) Supports(contentType, filename string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch ct {
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-excel.sheet.macroenabled.12",
		"application/vnd.ms-excel":
		return true
	}
	name := strings.ToLower(strings.TrimSpace(filename))
	for _, ext := range []string{".xlsx", ".xlsm", ".xls"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// Parse 将每个工作表解析为一个表格章节。
func (p *ExcelParser) Parse(ctx context.Context, in reportservice.ParseInput) (report.ParsedDocument, error) {
	if err := ctx.Err(); err != nil {
		return report.ParsedDocument{}, err
	}
	f, err := excelize.OpenReader(bytes.NewReader(in.Data))
	if err != nil {
		return report.ParsedDocument{}, err
	}
	defer func() { _ = f.Close() }()

	var (
		sections []report.ParsedSection
		raw      strings.Builder
	)
	for _, sheet := range f.GetSheetList() {
		if err := ctx.Err(); err != nil {
			return report.ParsedDocument{}, err
		}
		rows, err := f.GetRows(sheet)
		if err != nil {
			return report.ParsedDocument{}, err
		}
		headers, body := splitHeaderRows(rows)
		if len(headers) == 0 && len(body) == 0 {
			continue
		}
		table := report.NewTable(sheet, headers, body)
		md := table.Markdown()
		sections = append(sections, report.ParsedSection{
			Type:    report.SectionTable,
			Title:   sheet,
			Level:   1,
			Content: md,
			Tables:  []report.Table{table},
		})
		raw.WriteString(md)
		raw.WriteString("\n")
	}
	return report.ParsedDocument{
		RawText:  raw.String(),
		Sections: sections,
	}, nil
}

// splitHeaderRows 取第一个非空行为表头，其余非空行为数据。
func splitHeaderRows(rows [][]string) (headers []string, body [][]string) {
	headerTaken := false
	for _, row := range rows {
		if isEmptyRow(row) {
			continue
		}
		if !headerTaken {
			headers = append(headers, row...)
			headerTaken = true
			continue
		}
		body = append(body, row)
	}
	return headers, body
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
