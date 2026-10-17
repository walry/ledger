package report

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/idgen"
)

// Source 描述报告来源文件的元信息（值对象）。
type Source struct {
	Filename    string
	ContentType string
	Size        int64
	StoragePath string
}

// ParsedSection 是解析器产出的章节（值对象），由聚合根转换为 Section 实体。
type ParsedSection struct {
	Type    SectionType
	Title   string
	Level   int
	Content string
	Tables  []Table
}

// ParsedDocument 是解析器产出的文档结构。
type ParsedDocument struct {
	RawText  string
	Sections []ParsedSection
}

// Report 是财务报告聚合根。
//
// 不变量：
//   - id、name 必填
//   - 状态流转：uploaded → parsing → parsed/failed；parsed → archived
//   - 只有 parsed 状态可归档
type Report struct {
	id            ReportId
	name          string
	source        Source
	status        Status
	sections      []*Section
	rawText       string
	failureReason string
	uploadedAt    time.Time
	parsedAt      *time.Time
	archivedAt    *time.Time
	events        []ddd.DomainEvent
}

// NewReport 创建报告并记录 ReportUploaded 事件。
func NewReport(id ReportId, name string, source Source, now time.Time) (*Report, error) {
	if id.IsZero() {
		return nil, errors.New("report: id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("report: name is required")
	}
	r := &Report{
		id:         id,
		name:       name,
		source:     source,
		status:     StatusUploaded,
		sections:   make([]*Section, 0),
		uploadedAt: now,
	}
	r.record(NewReportUploaded(id, name, source, now))
	return r, nil
}

// Rehydrate 从持久化状态重建聚合，不产生新事件。
func Rehydrate(
	id ReportId, name string, source Source, status Status,
	sections []*Section, rawText, failureReason string,
	uploadedAt time.Time, parsedAt, archivedAt *time.Time,
) (*Report, error) {
	if id.IsZero() {
		return nil, errors.New("report: id is required")
	}
	if !status.Valid() {
		return nil, fmt.Errorf("report: invalid status %q", status)
	}
	if sections == nil {
		sections = make([]*Section, 0)
	}
	return &Report{
		id:            id,
		name:          name,
		source:        source,
		status:        status,
		sections:      sections,
		rawText:       rawText,
		failureReason: failureReason,
		uploadedAt:    uploadedAt,
		parsedAt:      parsedAt,
		archivedAt:    archivedAt,
	}, nil
}

// ID 返回聚合标识。
func (r *Report) ID() ReportId { return r.id }

// Name 返回报告名称。
func (r *Report) Name() string { return r.name }

// Source 返回来源元信息。
func (r *Report) Source() Source { return r.source }

// Status 返回当前状态。
func (r *Report) Status() Status { return r.status }

// Sections 返回章节列表（只读）。
func (r *Report) Sections() []*Section { return r.sections }

// RawText 返回解析出的全文。
func (r *Report) RawText() string { return r.rawText }

// FailureReason 返回失败原因。
func (r *Report) FailureReason() string { return r.failureReason }

// UploadedAt 返回上传时间。
func (r *Report) UploadedAt() time.Time { return r.uploadedAt }

// ParsedAt 返回解析完成时间。
func (r *Report) ParsedAt() *time.Time { return r.parsedAt }

// ArchivedAt 返回归档时间。
func (r *Report) ArchivedAt() *time.Time { return r.archivedAt }

// MarkParsing 将状态置为解析中。
func (r *Report) MarkParsing() error {
	if !r.status.CanParse() {
		return fmt.Errorf("report: cannot start parsing in status %s", r.status)
	}
	r.status = StatusParsing
	return nil
}

// MarkParsed 用解析结果填充聚合，并置为已解析。
func (r *Report) MarkParsed(doc ParsedDocument, ids idgen.Generator, now time.Time) error {
	if !r.status.CanParse() && r.status != StatusParsing {
		return fmt.Errorf("report: cannot finish parsing in status %s", r.status)
	}
	sections := make([]*Section, 0, len(doc.Sections))
	for i, ps := range doc.Sections {
		s := NewSection(SectionId(ids.New()), ps.Type, ps.Title, ps.Level, i)
		if ps.Content != "" {
			s.SetContent(ps.Content)
		}
		for _, t := range ps.Tables {
			s.AddTable(t)
		}
		sections = append(sections, s)
	}
	r.sections = sections
	r.rawText = doc.RawText
	r.status = StatusParsed
	r.failureReason = ""
	r.parsedAt = &now
	r.record(NewReportParsed(r.id, len(sections), now))
	return nil
}

// MarkFailed 置为解析失败。
func (r *Report) MarkFailed(reason string, now time.Time) {
	r.status = StatusFailed
	r.failureReason = reason
	r.record(NewReportParseFailed(r.id, reason, now))
}

// Archive 归档报告。
func (r *Report) Archive(now time.Time) error {
	if !r.status.CanArchive() {
		return fmt.Errorf("report: cannot archive in status %s", r.status)
	}
	r.status = StatusArchived
	r.archivedAt = &now
	return nil
}

// PullEvents 返回并清空累积的领域事件（由应用层调用后发布）。
func (r *Report) PullEvents() []ddd.DomainEvent {
	events := r.events
	r.events = nil
	return events
}

func (r *Report) record(e ddd.DomainEvent) { r.events = append(r.events, e) }
