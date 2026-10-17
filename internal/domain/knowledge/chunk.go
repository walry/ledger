package knowledge

import (
	"errors"
	"strings"
)

// Chunk 是文档聚合内的切片实体。
type Chunk struct {
	id         ChunkId
	index      int
	content    string
	embedding  Embedding
	metadata   map[string]string
	tokenCount int
}

// NewChunk 创建切片。
func NewChunk(id ChunkId, index int, content string, metadata map[string]string) (*Chunk, error) {
	if id.IsZero() {
		return nil, errors.New("knowledge: chunk id is required")
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("knowledge: chunk content is required")
	}
	if metadata == nil {
		metadata = map[string]string{}
	}
	return &Chunk{id: id, index: index, content: content, metadata: metadata}, nil
}

// RehydrateChunk 从持久化状态重建切片。
func RehydrateChunk(id ChunkId, index int, content string, emb Embedding, metadata map[string]string, tokenCount int) *Chunk {
	if metadata == nil {
		metadata = map[string]string{}
	}
	return &Chunk{id: id, index: index, content: content, embedding: emb, metadata: metadata, tokenCount: tokenCount}
}

// ID 返回切片标识。
func (c *Chunk) ID() ChunkId { return c.id }

// Index 返回切片在文档中的序号。
func (c *Chunk) Index() int { return c.index }

// Content 返回切片文本。
func (c *Chunk) Content() string { return c.content }

// Embedding 返回切片向量。
func (c *Chunk) Embedding() Embedding { return c.embedding }

// Metadata 返回元数据副本。
func (c *Chunk) Metadata() map[string]string {
	out := make(map[string]string, len(c.metadata))
	for k, v := range c.metadata {
		out[k] = v
	}
	return out
}

// TokenCount 返回估算的 token 数。
func (c *Chunk) TokenCount() int { return c.tokenCount }

// SetEmbedding 设置向量。
func (c *Chunk) SetEmbedding(e Embedding) { c.embedding = e }

// SetTokenCount 设置 token 估算值。
func (c *Chunk) SetTokenCount(n int) { c.tokenCount = n }
