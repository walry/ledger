package report

import (
	"context"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	reportdomain "github.com/walry/ledger/internal/domain/report"
)

// UploadReportUseCase 上传报告：创建聚合（记录 ReportUploaded）、保存并发布事件。
type UploadReportUseCase struct {
	repo    reportdomain.Repository
	factory *reportdomain.Factory
	bus     ddd.EventBus
}

// NewUploadReportUseCase 创建上传报告用例。
func NewUploadReportUseCase(repo reportdomain.Repository, factory *reportdomain.Factory, bus ddd.EventBus) *UploadReportUseCase {
	return &UploadReportUseCase{repo: repo, factory: factory, bus: bus}
}

// Execute 执行上传。
func (uc *UploadReportUseCase) Execute(ctx context.Context, cmd UploadReportCommand) (UploadReportResult, error) {
	source := reportdomain.Source{
		Filename:    cmd.Filename,
		ContentType: cmd.ContentType,
		Size:        int64(len(cmd.Data)),
		StoragePath: cmd.StoragePath,
	}
	r, err := uc.factory.Create(cmd.Name, source)
	if err != nil {
		return UploadReportResult{}, fmt.Errorf("report: create report: %w", err)
	}
	if err := uc.repo.Save(ctx, r); err != nil {
		return UploadReportResult{}, fmt.Errorf("report: save report: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, r.PullEvents()); err != nil {
		return UploadReportResult{}, fmt.Errorf("report: publish events: %w", err)
	}
	return UploadReportResult{ReportID: r.ID().String()}, nil
}
