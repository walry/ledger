package parser

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/walry/ledger/internal/domain/report"
	reportservice "github.com/walry/ledger/internal/domain/report/service"
)

// WordParser 实现 reportservice.Parser，仅用标准库解析 .docx。
//
// .docx 本质是一个 zip 包，正文位于 word/document.xml；本解析器只提取
// <w:t> 文本并依据 </w:p> 段落边界切分，不引入任何第三方 docx 依赖。
type WordParser struct{}

// NewWordParser 创建 Word 解析器。
func NewWordParser() *WordParser { return &WordParser{} }

// Supports 判断是否支持该格式。
func (p *WordParser) Supports(contentType, filename string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		return true
	}
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(filename)), ".docx")
}

// Parse 解析 docx 正文为段落章节。
func (p *WordParser) Parse(ctx context.Context, in reportservice.ParseInput) (report.ParsedDocument, error) {
	if err := ctx.Err(); err != nil {
		return report.ParsedDocument{}, err
	}
	zr, err := zip.NewReader(bytes.NewReader(in.Data), int64(len(in.Data)))
	if err != nil {
		return report.ParsedDocument{}, err
	}
	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return report.ParsedDocument{}, fmt.Errorf("parser: word/document.xml not found")
	}
	rc, err := doc.Open()
	if err != nil {
		return report.ParsedDocument{}, err
	}
	defer func() { _ = rc.Close() }()

	paras, err := extractDocxParagraphs(ctx, rc)
	if err != nil {
		return report.ParsedDocument{}, err
	}

	sections := make([]report.ParsedSection, 0, len(paras))
	var raw strings.Builder
	for _, p := range paras {
		if typ, level, ok := headingKind(p); ok {
			sections = append(sections, report.ParsedSection{Type: typ, Title: p, Level: level})
		} else {
			sections = append(sections, report.ParsedSection{Type: report.SectionParagraph, Content: p})
		}
		raw.WriteString(p)
		raw.WriteString("\n")
	}
	return report.ParsedDocument{RawText: raw.String(), Sections: sections}, nil
}

// extractDocxParagraphs 从 document.xml 中提取段落文本。
func extractDocxParagraphs(ctx context.Context, r io.Reader) ([]string, error) {
	dec := xml.NewDecoder(r)
	var (
		paras  []string
		para   strings.Builder
		inText bool
	)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "p":
				para.Reset()
			case "br", "cr":
				para.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				para.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				if s := strings.TrimSpace(para.String()); s != "" {
					paras = append(paras, s)
				}
				para.Reset()
			}
		}
	}
	if s := strings.TrimSpace(para.String()); s != "" {
		paras = append(paras, s)
	}
	return paras, nil
}
