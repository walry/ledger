// Package service 存放 report 上下文的领域服务与端口定义。
package service

import (
	"context"
	"errors"

	"github.com/walry/ledger/internal/domain/report"
)

// ErrUnsupportedFormat 表示没有解析器支持该格式。
var ErrUnsupportedFormat = errors.New("report: unsupported format")

// ParseInput 是解析器输入。
type ParseInput struct {
	Data        []byte
	Filename    string
	ContentType string
}

// Parser 是报告解析端口，实现在基础设施层（infrastructure/parser）。
//
// 端口方法仅使用领域语言（输入字节与文件名，输出领域文档结构），
// 不暴露任何具体解析库的类型。
type Parser interface {
	// Supports 判断当前解析器是否支持给定格式。
	Supports(contentType, filename string) bool
	// Parse 解析文件内容为领域文档结构。
	Parse(ctx context.Context, in ParseInput) (report.ParsedDocument, error)
}
