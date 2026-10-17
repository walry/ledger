// Package persistence wires up the GORM persistence adapters for every bounded
// context and exposes AutoMigrate for schema creation.
package persistence

import (
	"gorm.io/gorm"

	agentpersist "github.com/walry/ledger/internal/infrastructure/persistence/agent"
	analysispersist "github.com/walry/ledger/internal/infrastructure/persistence/analysis"
	conversationpersist "github.com/walry/ledger/internal/infrastructure/persistence/conversation"
	knowledgepersist "github.com/walry/ledger/internal/infrastructure/persistence/knowledge"
	modelconfigpersist "github.com/walry/ledger/internal/infrastructure/persistence/modelconfig"
	reportpersist "github.com/walry/ledger/internal/infrastructure/persistence/report"
	toolpersist "github.com/walry/ledger/internal/infrastructure/persistence/tool"
)

// AutoMigrate creates/updates all tables, foreign keys and indexes. Parents are
// listed before their children so foreign keys can be created.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// report
		&reportpersist.ReportRow{},
		&reportpersist.SectionRow{},

		// knowledge
		&knowledgepersist.KnowledgeBaseRow{},
		&knowledgepersist.DocumentRow{},
		&knowledgepersist.ChunkRow{},

		// agent
		&agentpersist.AgentRow{},

		// conversation
		&conversationpersist.ConversationRow{},
		&conversationpersist.TurnRow{},
		&conversationpersist.MessageRow{},

		// analysis
		&analysispersist.AnalysisTaskRow{},

		// tool
		&toolpersist.ToolRow{},

		// modelconfig
		&modelconfigpersist.ModelConfigRow{},
	)
}
