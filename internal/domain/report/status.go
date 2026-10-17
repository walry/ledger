package report

// Status 是报告的生命周期状态。
type Status string

const (
	// StatusUploaded 已上传，等待解析。
	StatusUploaded Status = "uploaded"
	// StatusParsing 解析中。
	StatusParsing Status = "parsing"
	// StatusParsed 解析完成。
	StatusParsed Status = "parsed"
	// StatusFailed 解析失败。
	StatusFailed Status = "failed"
	// StatusArchived 已归档。
	StatusArchived Status = "archived"
)

// Valid 判断状态是否为已知值。
func (s Status) Valid() bool {
	switch s {
	case StatusUploaded, StatusParsing, StatusParsed, StatusFailed, StatusArchived:
		return true
	default:
		return false
	}
}

// CanParse 判断当前状态是否允许开始解析。
func (s Status) CanParse() bool {
	return s == StatusUploaded || s == StatusFailed
}

// CanArchive 判断当前状态是否允许归档。
func (s Status) CanArchive() bool {
	return s == StatusParsed
}

// IsTerminal 判断是否为终态。
func (s Status) IsTerminal() bool {
	return s == StatusArchived
}

// String 返回字符串形式。
func (s Status) String() string { return string(s) }
