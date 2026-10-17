package knowledge

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// DocumentStatus 是文档索引状态。
type DocumentStatus string

const (
	// DocPending 待索引。
	DocPending DocumentStatus = "pending"
	// DocIndexed 已索引。
	DocIndexed DocumentStatus = "indexed"
	// DocFailed 索引失败。
	DocFailed DocumentStatus = "failed"
)

// Document 是文档聚合根，持有切片实体。
type Document struct {
	id            DocumentId
	kbID          KnowledgeBaseId
	source        SourceRef
	title         string
	chunks        []*Chunk
	status        DocumentStatus
	failureReason string
	createdAt     time.Time
	updatedAt     time.Time
	events        []ddd.DomainEvent
}

// NewDocument 创建文档。
func NewDocument(id DocumentId, kbID KnowledgeBaseId, source SourceRef, title string, now time.Time) (*Document, error) {
	if id.IsZero() {
		return nil, errors.New("knowledge: document id is required")
	}
	if kbID.IsZero() {
		return nil, errors.New("knowledge: knowledge base id is required")
	}
	if !source.Valid() {
		return nil, errors.New("knowledge: source ref is invalid")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = source.Title
	}
	return &Document{
		id:        id,
		kbID:      kbID,
		source:    source,
		title:     title,
		chunks:    make([]*Chunk, 0),
		status:    DocPending,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RehydrateDocument 从持久化状态重建文档。
func RehydrateDocument(
	id DocumentId, kbID KnowledgeBaseId, source SourceRef, title string,
	chunks []*Chunk, status DocumentStatus, failureReason string,
	createdAt, updatedAt time.Time,
) *Document {
	if chunks == nil {
		chunks = make([]*Chunk, 0)
	}
	return &Document{
		id: id, kbID: kbID, source: source, title: title, chunks: chunks,
		status: status, failureReason: failureReason, createdAt: createdAt, updatedAt: updatedAt,
	}
}

// ID 返回文档标识。
func (d *Document) ID() DocumentId { return d.id }

// KnowledgeBaseID 返回所属知识库标识。
func (d *Document) KnowledgeBaseID() KnowledgeBaseId { return d.kbID }

// Source 返回来源引用。
func (d *Document) Source() SourceRef { return d.source }

// Title 返回标题。
func (d *Document) Title() string { return d.title }

// Chunks 返回切片列表。
func (d *Document) Chunks() []*Chunk { return d.chunks }

// Status 返回索引状态。
func (d *Document) Status() DocumentStatus { return d.status }

// FailureReason 返回失败原因。
func (d *Document) FailureReason() string { return d.failureReason }

// CreatedAt 返回创建时间。
func (d *Document) CreatedAt() time.Time { return d.createdAt }

// UpdatedAt 返回更新时间。
func (d *Document) UpdatedAt() time.Time { return d.updatedAt }

// ReplaceChunks 用新的切片集合替换旧内容（用于重建索引），并重置为待索引。
func (d *Document) ReplaceChunks(chunks []*Chunk, now time.Time) error {
	if len(chunks) == 0 {
		return errors.New("knowledge: document must have at least one chunk")
	}
	d.chunks = chunks
	d.status = DocPending
	d.failureReason = ""
	d.updatedAt = now
	return nil
}

// MarkIndexed 标记索引完成并记录事件。
func (d *Document) MarkIndexed(now time.Time) error {
	if len(d.chunks) == 0 {
		return fmt.Errorf("knowledge: cannot mark indexed without chunks")
	}
	d.status = DocIndexed
	d.failureReason = ""
	d.updatedAt = now
	d.record(DocumentIndexed{
		KnowledgeBaseID: d.kbID,
		DocumentID:      d.id,
		ChunkCount:      len(d.chunks),
		Occurred:        now,
	})
	return nil
}

// MarkFailed 标记索引失败。
func (d *Document) MarkFailed(reason string, now time.Time) {
	d.status = DocFailed
	d.failureReason = reason
	d.updatedAt = now
}

// ToRef 返回聚合的 ID 引用。
func (d *Document) ToRef() DocumentRef {
	return DocumentRef{DocumentID: d.id, Title: d.title, ChunkCount: len(d.chunks)}
}

// PullEvents 返回并清空领域事件。
func (d *Document) PullEvents() []ddd.DomainEvent {
	events := d.events
	d.events = nil
	return events
}

func (d *Document) record(e ddd.DomainEvent) { d.events = append(d.events, e) }
