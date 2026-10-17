package knowledge

import (
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// KnowledgeBaseCreated 知识库已创建。
type KnowledgeBaseCreated struct {
	KnowledgeBaseID KnowledgeBaseId
	Name            string
	Occurred        time.Time
}

// EventName 实现 ddd.DomainEvent。
func (KnowledgeBaseCreated) EventName() string { return "KnowledgeBaseCreated" }

// OccurredAt 实现 ddd.DomainEvent。
func (e KnowledgeBaseCreated) OccurredAt() time.Time { return e.Occurred }

// DocumentIndexed 文档已建立索引。
type DocumentIndexed struct {
	KnowledgeBaseID KnowledgeBaseId
	DocumentID      DocumentId
	ChunkCount      int
	Occurred        time.Time
}

// EventName 实现 ddd.DomainEvent。
func (DocumentIndexed) EventName() string { return "DocumentIndexed" }

// OccurredAt 实现 ddd.DomainEvent。
func (e DocumentIndexed) OccurredAt() time.Time { return e.Occurred }

// DocumentRemoved 文档已移除。
type DocumentRemoved struct {
	KnowledgeBaseID KnowledgeBaseId
	DocumentID      DocumentId
	Occurred        time.Time
}

// EventName 实现 ddd.DomainEvent。
func (DocumentRemoved) EventName() string { return "DocumentRemoved" }

// OccurredAt 实现 ddd.DomainEvent。
func (e DocumentRemoved) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = KnowledgeBaseCreated{}
var _ ddd.DomainEvent = DocumentIndexed{}
var _ ddd.DomainEvent = DocumentRemoved{}
