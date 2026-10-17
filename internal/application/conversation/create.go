package conversation

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// CreateConversationUseCase 创建会话：解析智能体（需已发布）并绑定其当前版本。
type CreateConversationUseCase struct {
	convs  conversationdomain.Repository
	agents agentdomain.Repository
	ids    idgen.Generator
	clk    clock.Clock
	bus    ddd.EventBus
}

// NewCreateConversationUseCase 创建会话用例。
func NewCreateConversationUseCase(convs conversationdomain.Repository, agents agentdomain.Repository, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *CreateConversationUseCase {
	return &CreateConversationUseCase{convs: convs, agents: agents, ids: ids, clk: clk, bus: bus}
}

// Execute 执行创建。
func (uc *CreateConversationUseCase) Execute(ctx context.Context, cmd CreateConversationCommand) (CreateConversationResult, error) {
	a, err := uc.agents.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return CreateConversationResult{}, fmt.Errorf("conversation: agent %s not found: %w", cmd.AgentID, err)
		}
		return CreateConversationResult{}, fmt.Errorf("conversation: load agent: %w", err)
	}
	if a.Status() != agentdomain.StatusPublished {
		return CreateConversationResult{}, fmt.Errorf("conversation: agent %s is not published (status=%s)", cmd.AgentID, a.Status())
	}

	conv, err := conversationdomain.NewConversation(
		conversationdomain.ConversationId(uc.ids.New()),
		cmd.Title,
		a.ID(),
		a.Version(),
		uc.clk.Now(),
	)
	if err != nil {
		return CreateConversationResult{}, fmt.Errorf("conversation: create conversation: %w", err)
	}
	if err := uc.convs.Save(ctx, conv); err != nil {
		return CreateConversationResult{}, fmt.Errorf("conversation: save conversation: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, conv.PullEvents()); err != nil {
		return CreateConversationResult{}, fmt.Errorf("conversation: publish events: %w", err)
	}
	return CreateConversationResult{
		ConversationID: conv.ID().String(),
		AgentID:        conv.AgentID().String(),
		AgentVersion:   conv.AgentVersion(),
	}, nil
}
