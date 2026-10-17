// Package service 存放 conversation 上下文的领域服务与端口定义。
package service

import (
	"context"
	"errors"

	"github.com/walry/ledger/internal/domain/conversation"
)

// CompressedResult 是压缩结果（值对象）。
type CompressedResult struct {
	Summary  string
	Retained []conversation.Message
}

// CompressionStrategy 是上下文压缩策略端口。
//
// 实现必须逐字保留 UUID、路径、URL 等标识符，且不得把一次
// Tool 调用及其结果切散。
type CompressionStrategy interface {
	Compress(ctx context.Context, messages []conversation.Message, priorSummary string) (CompressedResult, error)
}

// ContextCompressor 是上下文压缩领域服务。
type ContextCompressor struct {
	strategy CompressionStrategy
	window   int
}

// NewContextCompressor 创建压缩服务。window 为保留的最近消息条数。
func NewContextCompressor(strategy CompressionStrategy, window int) *ContextCompressor {
	if window <= 0 {
		window = 20
	}
	return &ContextCompressor{strategy: strategy, window: window}
}

// Window 返回保留窗口大小。
func (c *ContextCompressor) Window() int { return c.window }

// ShouldCompress 判断是否需要压缩。
func (c *ContextCompressor) ShouldCompress(totalMessages int) bool {
	return c.strategy != nil && totalMessages > c.window
}

// Compress 执行压缩。当无法压缩（无策略或消息不足）时返回 ErrNoCompression。
func (c *ContextCompressor) Compress(ctx context.Context, messages []conversation.Message, priorSummary string) (CompressedResult, error) {
	if c.strategy == nil {
		return CompressedResult{}, ErrNoCompression
	}
	if len(messages) <= c.window {
		return CompressedResult{}, ErrNoCompression
	}
	head := messages[:len(messages)-c.window]
	tail := messages[len(messages)-c.window:]
	result, err := c.strategy.Compress(ctx, head, priorSummary)
	if err != nil {
		return CompressedResult{}, err
	}
	result.Retained = append(append([]conversation.Message{}, result.Retained...), tail...)
	return result, nil
}

// ErrNoCompression 表示无需压缩。
var ErrNoCompression = errors.New("conversation: no compression needed")
