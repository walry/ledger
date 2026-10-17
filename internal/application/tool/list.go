package tool

import (
	"context"
	"fmt"

	tooldomain "github.com/walry/ledger/internal/domain/tool"
)

// ListToolsUseCase 分页查询工具。
type ListToolsUseCase struct {
	repo tooldomain.Repository
}

// NewListToolsUseCase 创建工具列表用例。
func NewListToolsUseCase(repo tooldomain.Repository) *ListToolsUseCase {
	return &ListToolsUseCase{repo: repo}
}

// Execute 执行分页查询。
func (uc *ListToolsUseCase) Execute(ctx context.Context, cmd ListToolsCommand) (ListToolsResult, error) {
	tools, total, err := uc.repo.List(ctx, cmd.Offset, cmd.Limit)
	if err != nil {
		return ListToolsResult{}, fmt.Errorf("tool: list tools: %w", err)
	}
	items := make([]ToolSummaryDTO, 0, len(tools))
	for _, t := range tools {
		items = append(items, toToolSummary(t))
	}
	return ListToolsResult{Tools: items, Total: total}, nil
}

func toToolSummary(t *tooldomain.Tool) ToolSummaryDTO {
	return ToolSummaryDTO{
		ToolID:      t.ID().String(),
		Name:        t.Name(),
		Description: t.Description(),
		Enabled:     t.Enabled(),
		SourceKind:  string(t.Source().Kind()),
		Version:     t.Version(),
	}
}
