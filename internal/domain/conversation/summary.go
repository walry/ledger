package conversation

import (
	"errors"
	"strings"
	"time"
)

// ConversationSummary 是上下文压缩摘要（值对象）。
type ConversationSummary struct {
	content       string
	upToTurnIndex int
	createdAt     time.Time
}

// NewConversationSummary 创建摘要。
func NewConversationSummary(content string, upToTurnIndex int, now time.Time) (*ConversationSummary, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("conversation: summary content is required")
	}
	if upToTurnIndex < 0 {
		return nil, errors.New("conversation: summary upToTurnIndex must be non-negative")
	}
	return &ConversationSummary{content: content, upToTurnIndex: upToTurnIndex, createdAt: now}, nil
}

// Content 返回摘要文本。
func (s *ConversationSummary) Content() string { return s.content }

// UpToTurnIndex 返回摘要覆盖到的轮次序号（含）。
func (s *ConversationSummary) UpToTurnIndex() int { return s.upToTurnIndex }

// CreatedAt 返回创建时间。
func (s *ConversationSummary) CreatedAt() time.Time { return s.createdAt }
