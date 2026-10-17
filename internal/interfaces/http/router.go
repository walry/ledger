package http

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	agentapp "github.com/walry/ledger/internal/application/agent"
	analysisapp "github.com/walry/ledger/internal/application/analysis"
	conversationapp "github.com/walry/ledger/internal/application/conversation"
	knowledgeapp "github.com/walry/ledger/internal/application/knowledge"
	modelconfigapp "github.com/walry/ledger/internal/application/modelconfig"
	qaapp "github.com/walry/ledger/internal/application/qa"
	reportapp "github.com/walry/ledger/internal/application/report"
	toolapp "github.com/walry/ledger/internal/application/tool"
	"github.com/walry/ledger/internal/interfaces/http/middleware"
)

// Deps 汇总路由所需的全部用例与适配器。
type Deps struct {
	Logger *slog.Logger
	Files  FileStore

	UploadReport *reportapp.UploadReportUseCase
	ParseReport  *reportapp.ParseReportUseCase
	GetReport    *reportapp.GetReportUseCase
	ListReports  *reportapp.ListReportsUseCase

	CreateKnowledgeBase *knowledgeapp.CreateKnowledgeBaseUseCase
	AddDocument         *knowledgeapp.AddDocumentUseCase
	ReindexDocument     *knowledgeapp.ReindexDocumentUseCase
	SearchKnowledge     *knowledgeapp.SearchKnowledgeUseCase
	ListKnowledgeBases  *knowledgeapp.ListKnowledgeBasesUseCase
	GetKnowledgeBase    *knowledgeapp.GetKnowledgeBaseUseCase

	CreateDraftAgent *agentapp.CreateDraftUseCase
	AddSkill         *agentapp.AddSkillUseCase
	BindTool         *agentapp.BindToolUseCase
	PublishAgent     *agentapp.PublishUseCase
	NewAgentVersion  *agentapp.NewVersionUseCase
	RollbackAgent    *agentapp.RollbackUseCase
	ListAgents       *agentapp.ListAgentsUseCase
	GetAgent         *agentapp.GetAgentUseCase

	CreateConversation *conversationapp.CreateConversationUseCase
	StartTurn          *conversationapp.StartTurnUseCase
	CompressContext    *conversationapp.CompressContextUseCase
	ListConversations  *conversationapp.ListConversationsUseCase
	GetConversation    *conversationapp.GetConversationUseCase

	Ask       *qaapp.AskUseCase
	AskStream *qaapp.AskStreamUseCase

	SubmitTask *analysisapp.SubmitTaskUseCase
	RunTask    *analysisapp.RunTaskUseCase
	GetResult  *analysisapp.GetResultUseCase
	ListTasks  *analysisapp.ListTasksUseCase

	RegisterTool *toolapp.RegisterToolUseCase
	EnableTool   *toolapp.EnableToolUseCase
	ListTools    *toolapp.ListToolsUseCase
	GetTool      *toolapp.GetToolUseCase

	CreateModelConfig *modelconfigapp.CreateModelConfigUseCase
	UpdateModelConfig *modelconfigapp.UpdateModelConfigUseCase
	ListModelConfigs  *modelconfigapp.ListModelConfigsUseCase
	GetModelConfig    *modelconfigapp.GetModelConfigUseCase
}

// NewRouter 组装 HTTP 路由。
func NewRouter(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	logger := d.Logger
	if logger == nil {
		logger = slog.Default()
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware(), middleware.RecoveryMiddleware(logger), middleware.LogMiddleware(logger))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	reports := NewReportHandler(d.UploadReport, d.ParseReport, d.GetReport, d.ListReports, d.Files)
	kb := NewKnowledgeHandler(d.CreateKnowledgeBase, d.AddDocument, d.ReindexDocument, d.SearchKnowledge, d.ListKnowledgeBases, d.GetKnowledgeBase)
	agents := NewAgentHandler(d.CreateDraftAgent, d.AddSkill, d.BindTool, d.PublishAgent, d.NewAgentVersion, d.RollbackAgent, d.ListAgents, d.GetAgent)
	convs := NewConversationHandler(d.CreateConversation, d.StartTurn, d.CompressContext, d.ListConversations, d.GetConversation)
	qa := NewQAHandler(d.Ask, d.AskStream)
	analysis := NewAnalysisHandler(d.SubmitTask, d.RunTask, d.GetResult, d.ListTasks)
	tools := NewToolHandler(d.RegisterTool, d.EnableTool, d.ListTools, d.GetTool)
	models := NewModelConfigHandler(d.CreateModelConfig, d.UpdateModelConfig, d.ListModelConfigs, d.GetModelConfig)

	api := r.Group("/api")
	{
		api.POST("/reports", reports.Upload)
		api.GET("/reports", reports.List)
		api.GET("/reports/:id", reports.Get)
		api.POST("/reports/:id/parse", reports.Parse)

		api.GET("/knowledge-bases", kb.ListKnowledgeBases)
		api.POST("/knowledge-bases", kb.CreateKnowledgeBase)
		api.GET("/knowledge-bases/:id", kb.GetKnowledgeBase)
		api.POST("/knowledge-bases/:id/documents", kb.AddDocument)
		api.POST("/documents/:id/reindex", kb.Reindex)
		api.POST("/knowledge/search", kb.Search)

		api.GET("/agents", agents.List)
		api.GET("/agents/:id", agents.Get)
		api.POST("/agents", agents.CreateDraft)
		api.POST("/agents/:id/skills", agents.AddSkill)
		api.POST("/agents/:id/tools", agents.BindTool)
		api.POST("/agents/:id/publish", agents.Publish)
		api.POST("/agents/:id/versions", agents.NewVersion)
		api.POST("/agents/:id/rollback", agents.Rollback)

		api.GET("/conversations", convs.List)
		api.POST("/conversations", convs.Create)
		api.GET("/conversations/:id", convs.Get)
		api.POST("/conversations/:id/turns", convs.StartTurn)
		api.POST("/conversations/:id/compress", convs.Compress)

		api.POST("/qa/ask", qa.Ask)
		api.GET("/qa/ask/stream", qa.AskStream)
		api.POST("/qa/ask/stream", qa.AskStream)

		api.GET("/analysis/tasks", analysis.List)
		api.POST("/analysis/tasks", analysis.Submit)
		api.POST("/analysis/tasks/:id/run", analysis.Run)
		api.GET("/analysis/tasks/:id", analysis.Get)

		api.GET("/tools", tools.List)
		api.POST("/tools", tools.Register)
		api.GET("/tools/:id", tools.Get)
		api.POST("/tools/:id/enable", tools.Enable)

		api.GET("/model-configs", models.List)
		api.POST("/model-configs", models.Create)
		api.GET("/model-configs/:id", models.Get)
		api.PUT("/model-configs/:id", models.Update)
	}

	return r
}
