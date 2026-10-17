package knowledge

// SourceKind 描述文档来源类型。
type SourceKind string

const (
	// SourceReport 来源为财务报告。
	SourceReport SourceKind = "report"
	// SourceFile 来源为上传文件。
	SourceFile SourceKind = "file"
	// SourceURL 来源为外部链接。
	SourceURL SourceKind = "url"
)

// SourceRef 是对外部来源的 ID 引用（跨聚合只引用 ID，不引用对象）。
type SourceRef struct {
	Kind  SourceKind
	RefID string
	Title string
}

// Valid 校验来源引用。
func (s SourceRef) Valid() bool {
	return s.Kind != "" && s.RefID != ""
}

// DocumentRef 是知识库聚合对文档聚合的 ID 引用。
type DocumentRef struct {
	DocumentID DocumentId
	Title      string
	ChunkCount int
}
