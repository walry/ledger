package report

import (
	"context"
	"errors"
	"fmt"

	reportdomain "github.com/walry/ledger/internal/domain/report"
)

// GetReportUseCase 查询单个报告详情。
type GetReportUseCase struct {
	repo reportdomain.Repository
}

// NewGetReportUseCase 创建查询报告用例。
func NewGetReportUseCase(repo reportdomain.Repository) *GetReportUseCase {
	return &GetReportUseCase{repo: repo}
}

// Execute 执行查询。
func (uc *GetReportUseCase) Execute(ctx context.Context, cmd GetReportCommand) (GetReportResult, error) {
	id := reportdomain.NewReportId(cmd.ReportID)
	r, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, reportdomain.ErrNotFound) {
			return GetReportResult{}, fmt.Errorf("report: %s not found: %w", cmd.ReportID, err)
		}
		return GetReportResult{}, fmt.Errorf("report: load report: %w", err)
	}
	sections := make([]SectionDTO, 0, len(r.Sections()))
	for _, s := range r.Sections() {
		sections = append(sections, SectionDTO{
			ID:      s.ID().String(),
			Type:    string(s.Type()),
			Title:   s.Title(),
			Level:   s.Level(),
			Order:   s.Order(),
			Content: s.Content(),
		})
	}
	return GetReportResult{
		ReportID:      r.ID().String(),
		Name:          r.Name(),
		Status:        r.Status().String(),
		ContentType:   r.Source().ContentType,
		Filename:      r.Source().Filename,
		StoragePath:   r.Source().StoragePath,
		FailureReason: r.FailureReason(),
		SectionCount:  len(sections),
		Sections:      sections,
	}, nil
}
