// Package parser 提供 report 领域 Parser 端口的基础设施实现。
//
// 本包只依赖标准库以及 PDF/Excel 解析库，不感知应用层与运行时。
package parser

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/walry/ledger/internal/domain/report"
	reportservice "github.com/walry/ledger/internal/domain/report/service"
)

// PDFParser 实现 reportservice.Parser，基于 github.com/ledongthuc/pdf。
type PDFParser struct{}

// NewPDFParser 创建 PDF 解析器。
func NewPDFParser() *PDFParser { return &PDFParser{} }

var (
	reChapter   = regexp.MustCompile(`^第[一二三四五六七八九十百千零0-9]+章`)
	reSection   = regexp.MustCompile(`^第[一二三四五六七八九十百千零0-9]+节`)
	reCnNum     = regexp.MustCompile(`^[一二三四五六七八九十]+、`)
	reCnParen   = regexp.MustCompile(`^（[一二三四五六七八九十]+）`)
	reCnParenAs = regexp.MustCompile(`^\([一二三四五六七八九十]+\)`)
)

// Supports 判断是否支持该格式：content-type 为 application/pdf 或文件名为 .pdf。
func (p *PDFParser) Supports(contentType, filename string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "application/pdf" {
		return true
	}
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(filename)), ".pdf")
}

// Parse 提取 PDF 全文并按财务报告标题切分为章节。
func (p *PDFParser) Parse(ctx context.Context, in reportservice.ParseInput) (report.ParsedDocument, error) {
	if err := ctx.Err(); err != nil {
		return report.ParsedDocument{}, err
	}
	r, err := pdf.NewReader(bytes.NewReader(in.Data), int64(len(in.Data)))
	if err != nil {
		return report.ParsedDocument{}, err
	}
	text, err := extractPDFText(ctx, r)
	if err != nil {
		return report.ParsedDocument{}, err
	}
	return report.ParsedDocument{
		RawText:  text,
		Sections: splitSections(text),
	}, nil
}

func extractPDFText(ctx context.Context, r *pdf.Reader) (string, error) {
	rd, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	buf := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		n, err := rd.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

// splitSections 依据标题模式将纯文本切分为章节。
func splitSections(text string) []report.ParsedSection {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	sections := make([]report.ParsedSection, 0)
	var para []string
	flushPara := func() {
		content := strings.TrimSpace(strings.Join(para, "\n"))
		para = para[:0]
		if content == "" {
			return
		}
		sections = append(sections, report.ParsedSection{
			Type:    report.SectionParagraph,
			Content: content,
		})
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			flushPara()
			continue
		}
		if typ, level, ok := headingKind(line); ok {
			flushPara()
			sections = append(sections, report.ParsedSection{
				Type:  typ,
				Title: line,
				Level: level,
			})
			continue
		}
		para = append(para, line)
	}
	flushPara()
	return sections
}

// headingKind 识别一行是否为标题，并返回其类型与层级。
func headingKind(line string) (report.SectionType, int, bool) {
	switch {
	case reChapter.MatchString(line):
		return report.SectionChapter, 1, true
	case reSection.MatchString(line):
		return report.SectionHeading, 2, true
	case reCnParen.MatchString(line), reCnParenAs.MatchString(line):
		return report.SectionHeading, 3, true
	case reCnNum.MatchString(line):
		return report.SectionHeading, 2, true
	}
	// 全大写且较短的英文行视为标题。
	if isAllCapsHeading(line) {
		return report.SectionHeading, 2, true
	}
	return "", 0, false
}

func isAllCapsHeading(line string) bool {
	if len(line) < 2 || len(line) > 60 {
		return false
	}
	hasLetter := false
	for _, r := range line {
		switch {
		case r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= 'a' && r <= 'z':
			return false
		default:
			// 允许空格、数字与常见标点
		}
	}
	return hasLetter
}
