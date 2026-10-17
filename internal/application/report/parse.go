package report

import (
	"context"
	"errors"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	reportdomain "github.com/walry/ledger/internal/domain/report"
	reportservice "github.com/walry/ledger/internal/domain/report/service"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// ParseReportUseCase 解析报告：加载聚合、标记解析中、调用解析端口、写回解析结果。
type ParseReportUseCase struct {
	repo   reportdomain.Repository
	parser reportservice.Parser
	ids    idgen.Generator
	clk    clock.Clock
	bus    ddd.EventBus
}

// NewParseReportUseCase 创建解析报告用例。
func NewParseReportUseCase(repo reportdomain.Repository, parser reportservice.Parser, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *ParseReportUseCase {
	return &ParseReportUseCase{repo: repo, parser: parser, ids: ids, clk: clk, bus: bus}
}

// Execute 执行解析。
func (uc *ParseReportUseCase) Execute(ctx context.Context, cmd ParseReportCommand) (ParseReportResult, error) {
	id := reportdomain.NewReportId(cmd.ReportID)
	r, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, reportdomain.ErrNotFound) {
			return ParseReportResult{}, fmt.Errorf("report: %s not found: %w", cmd.ReportID, err)
		}
		return ParseReportResult{}, fmt.Errorf("report: load report: %w", err)
	}
	if err := r.MarkParsing(); err != nil {
		return ParseReportResult{}, fmt.Errorf("report: mark parsing: %w", err)
	}

	doc, parseErr := uc.parser.Parse(ctx, reportservice.ParseInput{
		Data:        cmd.Data,
		Filename:    cmd.Filename,
		ContentType: cmd.ContentType,
	})
	if parseErr != nil {
		r.MarkFailed(parseErr.Error(), uc.clk.Now())
		if err := uc.repo.Save(ctx, r); err != nil {
			return ParseReportResult{}, fmt.Errorf("report: save failed report: %w", err)
		}
		if err := publishEvents(ctx, uc.bus, r.PullEvents()); err != nil {
			return ParseReportResult{}, fmt.Errorf("report: publish events: %w", err)
		}
		return ParseReportResult{}, fmt.Errorf("report: parse: %w", parseErr)
	}

	if err := r.MarkParsed(doc, uc.ids, uc.clk.Now()); err != nil {
		return ParseReportResult{}, fmt.Errorf("report: mark parsed: %w", err)
	}
	if err := uc.repo.Save(ctx, r); err != nil {
		return ParseReportResult{}, fmt.Errorf("report: save report: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, r.PullEvents()); err != nil {
		return ParseReportResult{}, fmt.Errorf("report: publish events: %w", err)
	}
	return ParseReportResult{ReportID: r.ID().String(), SectionCount: len(r.Sections())}, nil
}
