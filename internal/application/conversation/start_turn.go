package conversation

import (
	"context"
	"errors"
	"fmt"

	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// StartTurnUseCase 开启一轮对话并追加用户消息。
type StartTurnUseCase struct {
	repo conversationdomain.Repository
	ids  idgen.Generator
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewStartTurnUseCase 创建开启轮次用例。
func NewStartTurnUseCase(repo conversationdomain.Repository, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *StartTurnUseCase {
	return &StartTurnUseCase{repo: repo, ids: ids, clk: clk, bus: bus}
}

// Execute 执行开启轮次。
func (uc *StartTurnUseCase) Execute(ctx context.Context, cmd StartTurnCommand) (StartTurnResult, error) {
	conv, err := uc.repo.FindByID(ctx, conversationdomain.NewConversationId(cmd.ConversationID))
	if err != nil {
		if errors.Is(err, conversationdomain.ErrNotFound) {
			return StartTurnResult{}, fmt.Errorf("conversation: %s not found: %w", cmd.ConversationID, err)
		}
		return StartTurnResult{}, fmt.Errorf("conversation: load conversation: %w", err)
	}
	now := uc.clk.Now()
	turn, err := conv.StartTurn(uc.ids, now)
	if err != nil {
		return StartTurnResult{}, fmt.Errorf("conversation: start turn: %w", err)
	}
	turn.AppendMessage(conversationdomain.NewUserMessage(cmd.Message, now))
	if err := uc.repo.Save(ctx, conv); err != nil {
		return StartTurnResult{}, fmt.Errorf("conversation: save conversation: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, conv.PullEvents()); err != nil {
		return StartTurnResult{}, fmt.Errorf("conversation: publish events: %w", err)
	}
	return StartTurnResult{ConversationID: conv.ID().String(), TurnID: turn.ID().String()}, nil
}
