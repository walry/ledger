package conversation

import (
	"context"
	"errors"

	"github.com/walry/ledger/internal/domain/agent"
)

// ErrNotFound 表示会话不存在。
var ErrNotFound = errors.New("conversation: not found")

// Repository 是会话聚合的仓储接口。
type Repository interface {
	Save(ctx context.Context, c *Conversation) error
	FindByID(ctx context.Context, id ConversationId) (*Conversation, error)
	ListByAgent(ctx context.Context, agentID agent.AgentId, offset, limit int) ([]*Conversation, int64, error)
	List(ctx context.Context, offset, limit int) ([]*Conversation, int64, error)
	Delete(ctx context.Context, id ConversationId) error
}
