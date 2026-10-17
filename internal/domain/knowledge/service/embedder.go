package service

import (
	"context"

	"github.com/walry/ledger/internal/domain/knowledge"
)

// Embedder 是向量化端口（实现在 infrastructure/embedding）。
type Embedder interface {
	// Embed 将一批文本转为向量，返回顺序与输入一致。
	Embed(ctx context.Context, texts []string) ([]knowledge.Embedding, error)
	// Dimension 返回向量维度。
	Dimension() int
	// Model 返回模型名。
	Model() string
}
