package tool

import (
	"context"
	"errors"
	"fmt"

	tooldomain "github.com/walry/ledger/internal/domain/tool"
)

// GetToolUseCase 查询单个工具详情。
type GetToolUseCase struct {
	repo tooldomain.Repository
}

// NewGetToolUseCase 创建工具详情用例。
func NewGetToolUseCase(repo tooldomain.Repository) *GetToolUseCase {
	return &GetToolUseCase{repo: repo}
}

// Execute 执行查询。
func (uc *GetToolUseCase) Execute(ctx context.Context, cmd GetToolCommand) (ToolDetailDTO, error) {
	t, err := uc.repo.FindByID(ctx, tooldomain.NewToolId(cmd.ToolID))
	if err != nil {
		if errors.Is(err, tooldomain.ErrNotFound) {
			return ToolDetailDTO{}, fmt.Errorf("tool: %s not found: %w", cmd.ToolID, err)
		}
		return ToolDetailDTO{}, fmt.Errorf("tool: load tool: %w", err)
	}
	src := t.Source()
	return ToolDetailDTO{
		ToolID:         t.ID().String(),
		Name:           t.Name(),
		Description:    t.Description(),
		Enabled:        t.Enabled(),
		Version:        t.Version(),
		SourceKind:     string(src.Kind()),
		SourceURI:      src.URI(),
		Handler:        t.Handler().Name(),
		Schema:         t.Schema(),
		TimeoutSeconds: int(t.Timeout().Seconds()),
		CreatedAt:      t.CreatedAt(),
		UpdatedAt:      t.UpdatedAt(),
	}, nil
}
