package report

import (
	"context"
	"fmt"

	reportdomain "github.com/walry/ledger/internal/domain/report"
)

// ListReportsUseCase 分页查询报告。
type ListReportsUseCase struct {
	repo reportdomain.Repository
}

// NewListReportsUseCase 创建报告列表用例。
func NewListReportsUseCase(repo reportdomain.Repository) *ListReportsUseCase {
	return &ListReportsUseCase{repo: repo}
}

// Execute 执行分页查询。
func (uc *ListReportsUseCase) Execute(ctx context.Context, cmd ListReportsCommand) (ListReportsResult, error) {
	var status reportdomain.Status
	if cmd.Status != "" {
		status = reportdomain.Status(cmd.Status)
	}
	reports, total, err := uc.repo.List(ctx, reportdomain.ListQuery{
		Status: status,
		Offset: cmd.Offset,
		Limit:  cmd.Limit,
	})
	if err != nil {
		return ListReportsResult{}, fmt.Errorf("report: list reports: %w", err)
	}
	items := make([]ReportSummaryDTO, 0, len(reports))
	for _, r := range reports {
		items = append(items, ReportSummaryDTO{
			ReportID:     r.ID().String(),
			Name:         r.Name(),
			Status:       r.Status().String(),
			SectionCount: len(r.Sections()),
		})
	}
	return ListReportsResult{Reports: items, Total: total}, nil
}
