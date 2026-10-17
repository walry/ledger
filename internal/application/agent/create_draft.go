package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	"github.com/walry/ledger/pkg/clock"
)

// CreateDraftUseCase 创建草稿智能体：校验并解析引用的模型配置后组装聚合。
type CreateDraftUseCase struct {
	repo    agentdomain.Repository
	factory *agentdomain.Factory
	models  modelconfigdomain.Repository
	clk     clock.Clock
	bus     ddd.EventBus
}

// NewCreateDraftUseCase 创建草稿智能体用例。
func NewCreateDraftUseCase(repo agentdomain.Repository, factory *agentdomain.Factory, models modelconfigdomain.Repository, clk clock.Clock, bus ddd.EventBus) *CreateDraftUseCase {
	return &CreateDraftUseCase{repo: repo, factory: factory, models: models, clk: clk, bus: bus}
}

// Execute 执行创建。
func (uc *CreateDraftUseCase) Execute(ctx context.Context, cmd CreateDraftCommand) (CreateDraftResult, error) {
	mc, err := uc.models.FindByID(ctx, modelconfigdomain.NewModelConfigId(cmd.ModelConfigID))
	if err != nil {
		if errors.Is(err, modelconfigdomain.ErrNotFound) {
			return CreateDraftResult{}, fmt.Errorf("agent: model config %s not found: %w", cmd.ModelConfigID, err)
		}
		return CreateDraftResult{}, fmt.Errorf("agent: load model config: %w", err)
	}
	ref, err := agentdomain.NewModelConfigRef(mc.ID(), mc.Model(), mc.Provider())
	if err != nil {
		return CreateDraftResult{}, fmt.Errorf("agent: build model ref: %w", err)
	}

	prompt, err := agentdomain.NewSystemPrompt(cmd.SystemPrompt)
	if err != nil {
		return CreateDraftResult{}, fmt.Errorf("agent: build system prompt: %w", err)
	}

	kbIDs := make([]knowledgedomain.KnowledgeBaseId, 0, len(cmd.KnowledgeBaseIDs))
	for _, id := range cmd.KnowledgeBaseIDs {
		kbIDs = append(kbIDs, knowledgedomain.NewKnowledgeBaseId(id))
	}
	actx, err := agentdomain.NewAgentContext(kbIDs, cmd.MaxSteps, cmd.RetrievalTopK)
	if err != nil {
		return CreateDraftResult{}, fmt.Errorf("agent: build context: %w", err)
	}

	a, err := uc.factory.Create(cmd.Name, cmd.Description, prompt, ref, actx)
	if err != nil {
		return CreateDraftResult{}, fmt.Errorf("agent: create agent: %w", err)
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return CreateDraftResult{}, fmt.Errorf("agent: save agent: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, a.PullEvents()); err != nil {
		return CreateDraftResult{}, fmt.Errorf("agent: publish events: %w", err)
	}
	return CreateDraftResult{AgentID: a.ID().String(), Version: a.Version()}, nil
}
