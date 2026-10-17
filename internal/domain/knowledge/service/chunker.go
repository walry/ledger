// Package service 存放 knowledge 上下文的领域服务与端口定义。
package service

import "context"

// ChunkOptions 是切片参数。
type ChunkOptions struct {
	MaxTokens int
	Overlap   int
}

// RawChunk 是切片器产出的原始切片。
type RawChunk struct {
	Index    int
	Content  string
	Metadata map[string]string
	Tokens   int
}

// Chunker 是文本切片端口。
type Chunker interface {
	Chunk(ctx context.Context, text string, opts ChunkOptions) ([]RawChunk, error)
}
