package report

import (
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// ReportUploaded 报告已上传。knowledge 上下文据此创建文档并建立索引。
type ReportUploaded struct {
	ReportID    ReportId
	Name        string
	ContentType string
	StoragePath string
	Occurred    time.Time
}

// NewReportUploaded 创建事件。
func NewReportUploaded(id ReportId, name string, source Source, now time.Time) ReportUploaded {
	return ReportUploaded{
		ReportID:    id,
		Name:        name,
		ContentType: source.ContentType,
		StoragePath: source.StoragePath,
		Occurred:    now,
	}
}

// EventName 实现 ddd.DomainEvent。
func (ReportUploaded) EventName() string { return "ReportUploaded" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ReportUploaded) OccurredAt() time.Time { return e.Occurred }

// ReportParsed 报告解析完成。
type ReportParsed struct {
	ReportID     ReportId
	SectionCount int
	Occurred     time.Time
}

// NewReportParsed 创建事件。
func NewReportParsed(id ReportId, sectionCount int, now time.Time) ReportParsed {
	return ReportParsed{ReportID: id, SectionCount: sectionCount, Occurred: now}
}

// EventName 实现 ddd.DomainEvent。
func (ReportParsed) EventName() string { return "ReportParsed" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ReportParsed) OccurredAt() time.Time { return e.Occurred }

// ReportParseFailed 报告解析失败。
type ReportParseFailed struct {
	ReportID ReportId
	Reason   string
	Occurred time.Time
}

// NewReportParseFailed 创建事件。
func NewReportParseFailed(id ReportId, reason string, now time.Time) ReportParseFailed {
	return ReportParseFailed{ReportID: id, Reason: reason, Occurred: now}
}

// EventName 实现 ddd.DomainEvent。
func (ReportParseFailed) EventName() string { return "ReportParseFailed" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ReportParseFailed) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = ReportUploaded{}
var _ ddd.DomainEvent = ReportParsed{}
var _ ddd.DomainEvent = ReportParseFailed{}
