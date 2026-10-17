package agent

import (
	"context"
	"errors"
)

// ErrNotFound 表示智能体不存在。
var ErrNotFound = errors.New("agent: not found")

// Repository 是智能体聚合的仓储接口。
type Repository interface {
	Save(ctx context.Context, a *Agent) error
	FindByID(ctx context.Context, id AgentId) (*Agent, error)
	// FindPublished 返回某个智能体族中最新已发布版本。
	FindPublished(ctx context.Context, id AgentId) (*Agent, error)
	List(ctx context.Context, offset, limit int) ([]*Agent, int64, error)
	Delete(ctx context.Context, id AgentId) error
}
