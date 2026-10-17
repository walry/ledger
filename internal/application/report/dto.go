package report

import (
	"context"

	"github.com/walry/ledger/internal/domain/ddd"
)

// UploadReportCommand 是上传报告的输入。
type UploadReportCommand struct {
	Name        string
	Filename    string
	ContentType string
	Data        []byte
	StoragePath string
}

// UploadReportResult 是上传报告的输出。
type UploadReportResult struct {
	ReportID string
}

// ParseReportCommand 是解析报告的输入。
type ParseReportCommand struct {
	ReportID    string
	Data        []byte
	Filename    string
	ContentType string
}

// ParseReportResult 是解析报告的输出。
type ParseReportResult struct {
	ReportID     string
	SectionCount int
}

// GetReportCommand 是查询单个报告的输入。
type GetReportCommand struct {
	ReportID string
}

// GetReportResult 是报告详情 DTO。
type GetReportResult struct {
	ReportID      string
	Name          string
	Status        string
	ContentType   string
	Filename      string
	StoragePath   string
	FailureReason string
	SectionCount  int
	Sections      []SectionDTO
}

// SectionDTO 是章节摘要 DTO。
type SectionDTO struct {
	ID      string
	Type    string
	Title   string
	Level   int
	Order   int
	Content string
}

// ListReportsCommand 是分页查询报告的输入。
type ListReportsCommand struct {
	Status string
	Offset int
	Limit  int
}

// ListReportsResult 是报告列表 DTO。
type ListReportsResult struct {
	Reports []ReportSummaryDTO
	Total   int64
}

// ReportSummaryDTO 是报告列表项 DTO。
type ReportSummaryDTO struct {
	ReportID     string
	Name         string
	Status       string
	SectionCount int
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
