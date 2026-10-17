package agent

import (
	"context"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// CreateDraftCommand 是创建草稿智能体的输入。
type CreateDraftCommand struct {
	Name             string
	Description      string
	SystemPrompt     string
	ModelConfigID    string
	KnowledgeBaseIDs []string
	MaxSteps         int
	RetrievalTopK    int
}

// CreateDraftResult 是创建草稿智能体的输出。
type CreateDraftResult struct {
	AgentID string
	Version int
}

// AddSkillCommand 是绑定技能的输入。
type AddSkillCommand struct {
	AgentID string
	SkillID string
	Version int
}

// AddSkillResult 是绑定技能的输出。
type AddSkillResult struct {
	AgentID string
}

// BindToolCommand 是绑定工具的输入。
type BindToolCommand struct {
	AgentID string
	ToolID  string
	Version int
	Config  map[string]string
}

// BindToolResult 是绑定工具的输出。
type BindToolResult struct {
	AgentID string
}

// PublishCommand 是发布智能体的输入。
type PublishCommand struct {
	AgentID string
}

// PublishResult 是发布智能体的输出。
type PublishResult struct {
	AgentID string
	Version int
}

// NewVersionCommand 是派生新版本的输入。
type NewVersionCommand struct {
	AgentID string
}

// NewVersionResult 是派生新版本的输出。
type NewVersionResult struct {
	AgentID       string
	Version       int
	ParentVersion int
}

// RollbackCommand 是回滚的输入（AgentID 为目标历史版本）。
type RollbackCommand struct {
	AgentID string
}

// RollbackResult 是回滚的输出。
type RollbackResult struct {
	AgentID       string
	Version       int
	ParentVersion int
}

// ListAgentsCommand 是分页查询智能体的输入。
type ListAgentsCommand struct {
	Offset int
	Limit  int
}

// ListAgentsResult 是智能体列表 DTO。
type ListAgentsResult struct {
	Agents []AgentSummaryDTO
	Total  int64
}

// AgentSummaryDTO 是智能体列表项 DTO。
type AgentSummaryDTO struct {
	AgentID       string
	Name          string
	Description   string
	Status        string
	Version       int
	ModelConfigID string
	Model         string
	Provider      string
}

// GetAgentCommand 是查询单个智能体的输入。
type GetAgentCommand struct {
	AgentID string
}

// AgentDetailDTO 是智能体详情 DTO。
type AgentDetailDTO struct {
	AgentID          string
	Name             string
	Description      string
	Status           string
	SystemPrompt     string
	ModelConfigID    string
	Model            string
	Provider         string
	KnowledgeBaseIDs []string
	MaxSteps         int
	RetrievalTopK    int
	Version          int
	ParentVersion    int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PublishedAt      *time.Time
	Skills           []SkillBindingDTO
	Tools            []ToolBindingDTO
}

// SkillBindingDTO 是技能绑定 DTO。
type SkillBindingDTO struct {
	SkillID string
	Version int
	Enabled bool
}

// ToolBindingDTO 是工具绑定 DTO。
type ToolBindingDTO struct {
	ToolID  string
	Version int
	Enabled bool
	Config  map[string]string
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
