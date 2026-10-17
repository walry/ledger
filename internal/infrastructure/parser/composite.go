package parser

import (
	"context"

	"github.com/walry/ledger/internal/domain/report"
	reportservice "github.com/walry/ledger/internal/domain/report/service"
)

var (
	_ reportservice.Parser = (*PDFParser)(nil)
	_ reportservice.Parser = (*ExcelParser)(nil)
	_ reportservice.Parser = (*WordParser)(nil)
	_ reportservice.Parser = (*TextParser)(nil)
	_ reportservice.Parser = (*CompositeParser)(nil)
)

// CompositeParser 按注册顺序委派给第一个 Supports 命中的解析器。
type CompositeParser struct {
	parsers []reportservice.Parser
}

// NewCompositeParser 创建组合解析器。
func NewCompositeParser(parsers ...reportservice.Parser) *CompositeParser {
	return &CompositeParser{parsers: parsers}
}

// NewDefaultParser 装配 PDF、Excel、Word、Text 四个解析器。
func NewDefaultParser() *CompositeParser {
	return NewCompositeParser(
		NewPDFParser(),
		NewExcelParser(),
		NewWordParser(),
		NewTextParser(),
	)
}

// Supports 判断是否有任一子解析器支持。
func (c *CompositeParser) Supports(contentType, filename string) bool {
	for _, p := range c.parsers {
		if p.Supports(contentType, filename) {
			return true
		}
	}
	return false
}

// Parse 委派给第一个支持的解析器；若无则返回 ErrUnsupportedFormat。
func (c *CompositeParser) Parse(ctx context.Context, in reportservice.ParseInput) (report.ParsedDocument, error) {
	for _, p := range c.parsers {
		if p.Supports(in.ContentType, in.Filename) {
			return p.Parse(ctx, in)
		}
	}
	return report.ParsedDocument{}, reportservice.ErrUnsupportedFormat
}
