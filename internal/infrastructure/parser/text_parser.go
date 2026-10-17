package parser

import (
	"context"
	"strings"

	"github.com/walry/ledger/internal/domain/report"
	reportservice "github.com/walry/ledger/internal/domain/report/service"
)

// TextParser 实现 reportservice.Parser，处理纯文本与 Markdown。
type TextParser struct{}

// NewTextParser 创建文本解析器。
func NewTextParser() *TextParser { return &TextParser{} }

// Supports 判断是否支持该格式。
func (p *TextParser) Supports(contentType, filename string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "text/plain" || ct == "text/markdown" || strings.HasPrefix(ct, "text/plain;") {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(filename))
	return strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".md")
}

// Parse 按空行切分为段落。
func (p *TextParser) Parse(ctx context.Context, in reportservice.ParseInput) (report.ParsedDocument, error) {
	if err := ctx.Err(); err != nil {
		return report.ParsedDocument{}, err
	}
	text := string(in.Data)
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	blocks := strings.Split(normalized, "\n\n")

	sections := make([]report.ParsedSection, 0, len(blocks))
	for _, block := range blocks {
		content := strings.TrimSpace(block)
		if content == "" {
			continue
		}
		sections = append(sections, report.ParsedSection{
			Type:    report.SectionParagraph,
			Content: content,
		})
	}
	return report.ParsedDocument{RawText: normalized, Sections: sections}, nil
}
