package knowledge

// KnowledgeBaseId 是知识库聚合根的标识。
type KnowledgeBaseId string

// NewKnowledgeBaseId 由原始字符串构造标识。
func NewKnowledgeBaseId(v string) KnowledgeBaseId { return KnowledgeBaseId(v) }

// String 返回字符串形式。
func (id KnowledgeBaseId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id KnowledgeBaseId) IsZero() bool { return id == "" }

// DocumentId 是文档聚合根的标识。
type DocumentId string

// NewDocumentId 由原始字符串构造标识。
func NewDocumentId(v string) DocumentId { return DocumentId(v) }

// String 返回字符串形式。
func (id DocumentId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id DocumentId) IsZero() bool { return id == "" }

// ChunkId 是切片实体（聚合内）的标识。
type ChunkId string

// NewChunkId 由原始字符串构造标识。
func NewChunkId(v string) ChunkId { return ChunkId(v) }

// String 返回字符串形式。
func (id ChunkId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id ChunkId) IsZero() bool { return id == "" }
