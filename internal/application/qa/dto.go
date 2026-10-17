package qa

// AskCommand 是问答助手的输入。
type AskCommand struct {
	ConversationID   string
	Question         string
	KnowledgeBaseIDs []string
}

// AskResult 是问答助手的输出。
type AskResult struct {
	Answer string
	TurnID string
	Tips   []string
}
