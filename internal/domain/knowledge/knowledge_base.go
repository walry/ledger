package knowledge

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// KnowledgeBase 是知识库聚合根。
//
// 它通过 DocumentRef 引用文档聚合（跨聚合只引用 ID，不引用对象）。
type KnowledgeBase struct {
	id             KnowledgeBaseId
	name           string
	description    string
	embeddingModel string
	dimension      int
	documents      []DocumentRef
	createdAt      time.Time
	updatedAt      time.Time
	events         []ddd.DomainEvent
}

// NewKnowledgeBase 创建知识库。
func NewKnowledgeBase(id KnowledgeBaseId, name, description, embeddingModel string, dimension int, now time.Time) (*KnowledgeBase, error) {
	if id.IsZero() {
		return nil, errors.New("knowledge: knowledge base id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("knowledge: knowledge base name is required")
	}
	kb := &KnowledgeBase{
		id:             id,
		name:           name,
		description:    description,
		embeddingModel: embeddingModel,
		dimension:      dimension,
		documents:      make([]DocumentRef, 0),
		createdAt:      now,
		updatedAt:      now,
	}
	kb.record(KnowledgeBaseCreated{KnowledgeBaseID: id, Name: name, Occurred: now})
	return kb, nil
}

// RehydrateKnowledgeBase 从持久化状态重建知识库。
func RehydrateKnowledgeBase(
	id KnowledgeBaseId, name, description, embeddingModel string, dimension int,
	documents []DocumentRef, createdAt, updatedAt time.Time,
) *KnowledgeBase {
	if documents == nil {
		documents = make([]DocumentRef, 0)
	}
	return &KnowledgeBase{
		id: id, name: name, description: description, embeddingModel: embeddingModel,
		dimension: dimension, documents: documents, createdAt: createdAt, updatedAt: updatedAt,
	}
}

// ID 返回聚合标识。
func (k *KnowledgeBase) ID() KnowledgeBaseId { return k.id }

// Name 返回名称。
func (k *KnowledgeBase) Name() string { return k.name }

// Description 返回描述。
func (k *KnowledgeBase) Description() string { return k.description }

// EmbeddingModel 返回使用的向量模型名。
func (k *KnowledgeBase) EmbeddingModel() string { return k.embeddingModel }

// Dimension 返回向量维度。
func (k *KnowledgeBase) Dimension() int { return k.dimension }

// Documents 返回文档引用列表。
func (k *KnowledgeBase) Documents() []DocumentRef {
	out := make([]DocumentRef, len(k.documents))
	copy(out, k.documents)
	return out
}

// CreatedAt 返回创建时间。
func (k *KnowledgeBase) CreatedAt() time.Time { return k.createdAt }

// UpdatedAt 返回更新时间。
func (k *KnowledgeBase) UpdatedAt() time.Time { return k.updatedAt }

// AddDocument 登记一个文档引用（幂等：同 ID 覆盖）。
func (k *KnowledgeBase) AddDocument(ref DocumentRef, now time.Time) error {
	if ref.DocumentID.IsZero() {
		return errors.New("knowledge: document ref id is required")
	}
	for i, existing := range k.documents {
		if existing.DocumentID == ref.DocumentID {
			k.documents[i] = ref
			k.updatedAt = now
			return nil
		}
	}
	k.documents = append(k.documents, ref)
	k.updatedAt = now
	return nil
}

// RemoveDocument 移除文档引用。
func (k *KnowledgeBase) RemoveDocument(id DocumentId, now time.Time) error {
	idx := -1
	for i, existing := range k.documents {
		if existing.DocumentID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("knowledge: document %s not in knowledge base", id)
	}
	k.documents = append(k.documents[:idx], k.documents[idx+1:]...)
	k.updatedAt = now
	k.record(DocumentRemoved{KnowledgeBaseID: k.id, DocumentID: id, Occurred: now})
	return nil
}

// PullEvents 返回并清空领域事件。
func (k *KnowledgeBase) PullEvents() []ddd.DomainEvent {
	events := k.events
	k.events = nil
	return events
}

func (k *KnowledgeBase) record(e ddd.DomainEvent) { k.events = append(k.events, e) }
