package conversation

import (
	"context"
	"errors"
	"fmt"

	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	convsrv "github.com/walry/ledger/internal/domain/conversation/service"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/clock"
)

// CompressContextUseCase 压缩会话上下文：调用压缩领域服务并回写摘要。
type CompressContextUseCase struct {
	repo       conversationdomain.Repository
	compressor *convsrv.ContextCompressor
	clk        clock.Clock
	bus        ddd.EventBus
}

// NewCompressContextUseCase 创建上下文压缩用例。
func NewCompressContextUseCase(repo conversationdomain.Repository, compressor *convsrv.ContextCompressor, clk clock.Clock, bus ddd.EventBus) *CompressContextUseCase {
	return &CompressContextUseCase{repo: repo, compressor: compressor, clk: clk, bus: bus}
}

// Execute 执行压缩。
func (uc *CompressContextUseCase) Execute(ctx context.Context, cmd CompressContextCommand) (CompressContextResult, error) {
	conv, err := uc.repo.FindByID(ctx, conversationdomain.NewConversationId(cmd.ConversationID))
	if err != nil {
		if errors.Is(err, conversationdomain.ErrNotFound) {
			return CompressContextResult{}, fmt.Errorf("conversation: %s not found: %w", cmd.ConversationID, err)
		}
		return CompressContextResult{}, fmt.Errorf("conversation: load conversation: %w", err)
	}

	msgs := make([]conversationdomain.Message, 0)
	for _, t := range conv.Turns() {
		msgs = append(msgs, t.Messages()...)
	}
	prior := ""
	if s := conv.Summary(); s != nil {
		prior = s.Content()
	}

	result, cerr := uc.compressor.Compress(ctx, msgs, prior)
	if errors.Is(cerr, convsrv.ErrNoCompression) {
		return CompressContextResult{ConversationID: conv.ID().String(), Compressed: false}, nil
	}
	if cerr != nil {
		return CompressContextResult{}, fmt.Errorf("conversation: compress: %w", cerr)
	}

	upTo := len(conv.Turns()) - 1
	if upTo < 0 {
		upTo = 0
	}
	summary, err := conversationdomain.NewConversationSummary(result.Summary, upTo, uc.clk.Now())
	if err != nil {
		return CompressContextResult{}, fmt.Errorf("conversation: build summary: %w", err)
	}
	conv.SetSummary(summary, uc.clk.Now())
	if err := uc.repo.Save(ctx, conv); err != nil {
		return CompressContextResult{}, fmt.Errorf("conversation: save conversation: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, conv.PullEvents()); err != nil {
		return CompressContextResult{}, fmt.Errorf("conversation: publish events: %w", err)
	}
	return CompressContextResult{
		ConversationID: conv.ID().String(),
		Summary:        summary.Content(),
		Compressed:     true,
	}, nil
}
