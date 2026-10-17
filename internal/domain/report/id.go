package report

// ReportId 是财务报告聚合根的标识。
type ReportId string

// NewReportId 由原始字符串构造标识。
func NewReportId(v string) ReportId { return ReportId(v) }

// String 返回字符串形式。
func (id ReportId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id ReportId) IsZero() bool { return id == "" }

// SectionId 是报告内章节实体（聚合内）的标识。
type SectionId string

// NewSectionId 由原始字符串构造标识。
func NewSectionId(v string) SectionId { return SectionId(v) }

// String 返回字符串形式。
func (id SectionId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id SectionId) IsZero() bool { return id == "" }
