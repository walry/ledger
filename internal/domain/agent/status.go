package agent

// Status 是智能体的生命周期状态。
type Status string

const (
	// StatusDraft 草稿，可编辑。
	StatusDraft Status = "draft"
	// StatusPublished 已发布，不可编辑。
	StatusPublished Status = "published"
	// StatusArchived 已归档。
	StatusArchived Status = "archived"
)

// Valid 判断状态是否为已知值。
func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusArchived:
		return true
	default:
		return false
	}
}

// Editable 判断当前状态是否允许修改配置。
func (s Status) Editable() bool { return s == StatusDraft }

// String 返回字符串形式。
func (s Status) String() string { return string(s) }
