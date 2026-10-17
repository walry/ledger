package analysis

// AnalysisTaskId 是分析任务聚合根的标识。
type AnalysisTaskId string

// NewAnalysisTaskId 由原始字符串构造标识。
func NewAnalysisTaskId(v string) AnalysisTaskId { return AnalysisTaskId(v) }

// String 返回字符串形式。
func (id AnalysisTaskId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id AnalysisTaskId) IsZero() bool { return id == "" }
