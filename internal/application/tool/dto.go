package tool

import (
	"context"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
)

// RegisterToolCommand 是注册工具的输入。
type RegisterToolCommand struct {
	Name           string
	Description    string
	HandlerName    string
	Schema         tooldomain.JSONSchema
	SourceKind     tooldomain.SourceKind
	SourceURI      string
	SourceConfig   map[string]string
	TimeoutSeconds int
}

// RegisterToolResult 是注册工具的输出。
type RegisterToolResult struct {
	ToolID string
}

// EnableToolCommand 是启用/停用工具的输入。
type EnableToolCommand struct {
	ToolID  string
	Enabled bool
}

// EnableToolResult 是启用/停用工具的输出。
type EnableToolResult struct {
	ToolID  string
	Enabled bool
}

// ListToolsCommand 是分页查询工具的输入。
type ListToolsCommand struct {
	Offset int
	Limit  int
}

// ListToolsResult 是工具列表 DTO。
type ListToolsResult struct {
	Tools []ToolSummaryDTO
	Total int64
}

// ToolSummaryDTO 是工具列表项 DTO。
type ToolSummaryDTO struct {
	ToolID      string
	Name        string
	Description string
	Enabled     bool
	SourceKind  string
	Version     int
}

// GetToolCommand 是查询单个工具的输入。
type GetToolCommand struct {
	ToolID string
}

// ToolDetailDTO 是工具详情 DTO。
type ToolDetailDTO struct {
	ToolID         string
	Name           string
	Description    string
	Enabled        bool
	Version        int
	SourceKind     string
	SourceURI      string
	Handler        string
	Schema         tooldomain.JSONSchema
	TimeoutSeconds int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
