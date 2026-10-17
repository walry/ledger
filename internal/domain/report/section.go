package report

// SectionType 是章节的类型（值对象枚举）。
type SectionType string

const (
	// SectionChapter 章。
	SectionChapter SectionType = "chapter"
	// SectionHeading 节/标题。
	SectionHeading SectionType = "heading"
	// SectionParagraph 正文段落。
	SectionParagraph SectionType = "paragraph"
	// SectionTable 表格。
	SectionTable SectionType = "table"
	// SectionNote 附注。
	SectionNote SectionType = "note"
)

// Section 是报告聚合内的实体（一致性边界之内，随聚合根一起持久化）。
type Section struct {
	id      SectionId
	typ     SectionType
	title   string
	level   int
	content string
	order   int
	tables  []Table
}

// NewSection 创建章节。仅供聚合根与解析器使用。
func NewSection(id SectionId, typ SectionType, title string, level, order int) *Section {
	if typ == "" {
		typ = SectionParagraph
	}
	return &Section{
		id:     id,
		typ:    typ,
		title:  title,
		level:  level,
		order:  order,
		tables: make([]Table, 0),
	}
}

// ID 返回章节标识。
func (s *Section) ID() SectionId { return s.id }

// Type 返回章节类型。
func (s *Section) Type() SectionType { return s.typ }

// Title 返回章节标题。
func (s *Section) Title() string { return s.title }

// Level 返回层级（如标题级别）。
func (s *Section) Level() int { return s.level }

// Content 返回正文内容。
func (s *Section) Content() string { return s.content }

// Order 返回在报告中的顺序。
func (s *Section) Order() int { return s.order }

// Tables 返回章节内表格的副本。
func (s *Section) Tables() []Table {
	out := make([]Table, len(s.tables))
	copy(out, s.tables)
	return out
}

// SetContent 设置正文内容。
func (s *Section) SetContent(content string) { s.content = content }

// AddTable 向章节追加一个表格。
func (s *Section) AddTable(t Table) { s.tables = append(s.tables, t) }

// Text 返回章节的纯文本表示（标题 + 正文 + 表格 Markdown），
// 用于切片、向量化与提示词上下文。
func (s *Section) Text() string {
	var b []byte
	if s.title != "" {
		b = append(b, s.title...)
		b = append(b, '\n')
	}
	if s.content != "" {
		b = append(b, s.content...)
		b = append(b, '\n')
	}
	for _, t := range s.tables {
		if t.Empty() {
			continue
		}
		b = append(b, t.Markdown()...)
	}
	return string(b)
}
