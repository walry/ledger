package knowledge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	reportdomain "github.com/walry/ledger/internal/domain/report"
)

// NewReportUploadedHandler 构造订阅 ReportUploaded 的事件处理器：
// 报告上传后，将报告全文抽取为文档并索引到指定知识库。
//
// 幂等：若同一 (kbID, SourceReport, reportID) 的文档已存在则直接返回。
// 订阅关系（bus.Subscribe("ReportUploaded", handler)）在装配层注册。
func NewReportUploadedHandler(uc *AddDocumentUseCase, reportRepo reportdomain.Repository, kbID knowledgedomain.KnowledgeBaseId) ddd.Handler {
	return func(ctx context.Context, event ddd.DomainEvent) error {
		ev, ok := event.(reportdomain.ReportUploaded)
		if !ok {
			if ptr, ptrOK := event.(*reportdomain.ReportUploaded); ptrOK && ptr != nil {
				ev = *ptr
			} else {
				return nil
			}
		}
		if ev.ReportID.IsZero() {
			return nil
		}

		existing, err := uc.docs.FindBySource(ctx, kbID, knowledgedomain.SourceReport, ev.ReportID.String())
		if err == nil && existing != nil {
			return nil
		}
		if err != nil && !errors.Is(err, knowledgedomain.ErrNotFound) {
			return fmt.Errorf("knowledge: check existing document: %w", err)
		}

		r, err := reportRepo.FindByID(ctx, ev.ReportID)
		if err != nil {
			return fmt.Errorf("knowledge: load report %s: %w", ev.ReportID, err)
		}
		text := reportText(r)
		if strings.TrimSpace(text) == "" {
			return nil
		}

		_, err = uc.Execute(ctx, AddDocumentCommand{
			KnowledgeBaseID: kbID.String(),
			Title:           r.Name(),
			SourceKind:      knowledgedomain.SourceReport,
			SourceRefID:     r.ID().String(),
			Text:            text,
		})
		return err
	}
}

// reportText 返回报告的纯文本（优先全文，否则拼接章节文本）。
func reportText(r *reportdomain.Report) string {
	if raw := strings.TrimSpace(r.RawText()); raw != "" {
		return raw
	}
	var b strings.Builder
	for _, s := range r.Sections() {
		b.WriteString(s.Text())
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
