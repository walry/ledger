// Command api 是财报分析助手应用的装配入口。
//
// 装配层是唯一知道"谁实现谁、谁订阅谁"的地方：
// 在此创建基础设施、领域服务、应用用例与接口层，并完成依赖注入与事件订阅。
package main

import (
	"context"
	"errors"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	agentapp "github.com/walry/ledger/internal/application/agent"
	analysisapp "github.com/walry/ledger/internal/application/analysis"
	conversationapp "github.com/walry/ledger/internal/application/conversation"
	knowledgeapp "github.com/walry/ledger/internal/application/knowledge"
	modelconfigapp "github.com/walry/ledger/internal/application/modelconfig"
	qaapp "github.com/walry/ledger/internal/application/qa"
	reportapp "github.com/walry/ledger/internal/application/report"
	toolapp "github.com/walry/ledger/internal/application/tool"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	convsrv "github.com/walry/ledger/internal/domain/conversation/service"
	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	reportdomain "github.com/walry/ledger/internal/domain/report"
	runtimedomain "github.com/walry/ledger/internal/domain/runtime"
	tooldomain "github.com/walry/ledger/internal/domain/tool"

	"github.com/walry/ledger/internal/infrastructure/chunker"
	"github.com/walry/ledger/internal/infrastructure/config"
	"github.com/walry/ledger/internal/infrastructure/embedding"
	"github.com/walry/ledger/internal/infrastructure/eventbus"
	"github.com/walry/ledger/internal/infrastructure/filestore"
	infraidgen "github.com/walry/ledger/internal/infrastructure/idgen"
	"github.com/walry/ledger/internal/infrastructure/llm"
	"github.com/walry/ledger/internal/infrastructure/parser"
	"github.com/walry/ledger/internal/infrastructure/persistence"
	persistenceagent "github.com/walry/ledger/internal/infrastructure/persistence/agent"
	persistenceanalysis "github.com/walry/ledger/internal/infrastructure/persistence/analysis"
	persistenceconversation "github.com/walry/ledger/internal/infrastructure/persistence/conversation"
	persistenceknowledge "github.com/walry/ledger/internal/infrastructure/persistence/knowledge"
	persistencemodelconfig "github.com/walry/ledger/internal/infrastructure/persistence/modelconfig"
	persistencereport "github.com/walry/ledger/internal/infrastructure/persistence/report"
	persistencetool "github.com/walry/ledger/internal/infrastructure/persistence/tool"
	"github.com/walry/ledger/internal/infrastructure/toolhandler"
	"github.com/walry/ledger/internal/infrastructure/vectorstore"
	httpiface "github.com/walry/ledger/internal/interfaces/http"
	"github.com/walry/ledger/internal/interfaces/job"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. 基础设施：数据库
	if cfg.DatabaseDSN == "" {
		return errors.New("DATABASE_DSN is required (postgres dsn)")
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		return err
	}
	if cfg.AutoMigrate {
		if err := persistence.AutoMigrate(db); err != nil {
			return err
		}
	}

	// 2. 标识符与时间
	ids := idgen.Generator(infraidgen.NewUUIDGenerator())
	clk := clock.System()

	// 3. 事件总线
	bus := eventbus.NewInMemoryBus()

	// 4. 仓储
	reportRepo := persistencereport.NewRepository(db)
	kbStore := persistenceknowledge.NewRepository(db)
	docRepo := kbStore.Documents()
	agentRepo := persistenceagent.NewRepository(db)
	convRepo := persistenceconversation.NewRepository(db)
	analysisRepo := persistenceanalysis.NewRepository(db)
	toolRepo := persistencetool.NewRepository(db)
	modelConfigRepo := persistencemodelconfig.NewRepository(db)

	// 5. 技术适配器（端口实现）
	parserAdapter := parser.NewDefaultParser()
	chunkerAdapter := chunker.NewRecursiveChunker()
	llmFactory := llm.NewClientFactory()

	embeddingModel := cfg.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
	embedder := embedding.NewOpenAIEmbedder(cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, embeddingModel, cfg.EmbeddingDimension)
	vectorStore := vectorstore.NewPgVectorStore(db, embedder, cfg.VectorDimension)
	if err := vectorStore.EnsureSchema(ctx); err != nil {
		logger.Warn("ensure vector schema failed; retrieval may not work", "error", err)
	}
	reranker := vectorstore.NewPassthroughReranker()

	toolRegistry := toolhandler.NewRegistry()
	toolhandler.RegisterBuiltins(toolRegistry, toolhandler.BuiltinConfig{
		WebSearchEndpoint: cfg.WebSearchEndpoint,
		WebSearchAPIKey:   cfg.WebSearchAPIKey,
		SQLDSN:            cfg.SQLDSN,
	})

	files, err := filestore.NewLocalStore(os.Getenv("UPLOAD_DIR"))
	if err != nil {
		return err
	}

	// 6. 领域服务
	reportFactory := reportdomain.NewFactory(ids, clk)
	agentFactory := agentdomain.NewFactory(ids, clk)

	compressor := convsrv.NewContextCompressor(convsrv.TruncationStrategy{}, 20)
	if client, model := defaultModelClient(llmFactory, cfg); client != nil {
		summarizer := convsrv.NewSummarizingStrategy(client, model)
		compressor = convsrv.NewContextCompressor(convsrv.NewSlidingWindowSummaryStrategy(summarizer), 20)
	} else {
		logger.Warn("no default LLM configured; context compression falls back to truncation only")
	}

	rag := runtimedomain.NewRAGPipeline(vectorStore, reranker, 5)

	executor := runtimedomain.NewAgentExecutor(runtimedomain.AgentExecutorConfig{
		Agents:        agentRepo,
		Conversations: convRepo,
		Tools:         toolRepo,
		ToolFactory:   toolRegistry,
		Models:        llmFactory,
		ModelConfigs:  modelConfigRepo,
		RAG:           rag,
		Compressor:    compressor,
		Events:        bus,
		IDs:           ids,
		Clock:         clk,
	})

	employee := runtimedomain.NewDigitalEmployee(runtimedomain.DigitalEmployeeConfig{
		Executor:      executor,
		Tasks:         analysisRepo,
		Reports:       reportRepo,
		Conversations: convRepo,
		Events:        bus,
		IDs:           ids,
		Clock:         clk,
	})

	qaAssistant := runtimedomain.NewQAAssistant(executor)

	// 7. 应用用例
	uploadUC := reportapp.NewUploadReportUseCase(reportRepo, reportFactory, bus)
	parseUC := reportapp.NewParseReportUseCase(reportRepo, parserAdapter, ids, clk, bus)
	getReportUC := reportapp.NewGetReportUseCase(reportRepo)
	listReportsUC := reportapp.NewListReportsUseCase(reportRepo)

	createKBUC := knowledgeapp.NewCreateKnowledgeBaseUseCase(kbStore, ids, clk, bus)
	addDocUC := knowledgeapp.NewAddDocumentUseCase(kbStore, docRepo, chunkerAdapter, embedder, vectorStore, ids, clk, bus)
	reindexUC := knowledgeapp.NewReindexDocumentUseCase(kbStore, docRepo, chunkerAdapter, embedder, vectorStore, ids, clk, bus)
	searchUC := knowledgeapp.NewSearchKnowledgeUseCase(vectorStore)
	listKBUC := knowledgeapp.NewListKnowledgeBasesUseCase(kbStore)
	getKBUC := knowledgeapp.NewGetKnowledgeBaseUseCase(kbStore, docRepo)

	createDraftUC := agentapp.NewCreateDraftUseCase(agentRepo, agentFactory, modelConfigRepo, clk, bus)
	addSkillUC := agentapp.NewAddSkillUseCase(agentRepo, clk, bus)
	bindToolUC := agentapp.NewBindToolUseCase(agentRepo, toolRepo, clk, bus)
	publishUC := agentapp.NewPublishUseCase(agentRepo, clk, bus)
	newVersionUC := agentapp.NewNewVersionUseCase(agentRepo, agentFactory, bus)
	rollbackUC := agentapp.NewRollbackUseCase(agentRepo, agentFactory, bus)
	listAgentsUC := agentapp.NewListAgentsUseCase(agentRepo)
	getAgentUC := agentapp.NewGetAgentUseCase(agentRepo)

	createConvUC := conversationapp.NewCreateConversationUseCase(convRepo, agentRepo, ids, clk, bus)
	startTurnUC := conversationapp.NewStartTurnUseCase(convRepo, ids, clk, bus)
	compressUC := conversationapp.NewCompressContextUseCase(convRepo, compressor, clk, bus)
	listConvsUC := conversationapp.NewListConversationsUseCase(convRepo)
	getConvUC := conversationapp.NewGetConversationUseCase(convRepo)

	askUC := qaapp.NewAskUseCase(qaAssistant)
	askStreamUC := qaapp.NewAskStreamUseCase(qaAssistant)

	submitTaskUC := analysisapp.NewSubmitTaskUseCase(analysisRepo, agentRepo, ids, clk, bus)
	runTaskUC := analysisapp.NewRunTaskUseCase(employee)
	getResultUC := analysisapp.NewGetResultUseCase(analysisRepo)
	listTasksUC := analysisapp.NewListTasksUseCase(analysisRepo)

	registerToolUC := toolapp.NewRegisterToolUseCase(toolRepo, toolRegistry, ids, clk, bus)
	enableToolUC := toolapp.NewEnableToolUseCase(toolRepo, clk, bus)
	listToolsUC := toolapp.NewListToolsUseCase(toolRepo)
	getToolUC := toolapp.NewGetToolUseCase(toolRepo)

	createModelUC := modelconfigapp.NewCreateModelConfigUseCase(modelConfigRepo, ids, clk, bus)
	updateModelUC := modelconfigapp.NewUpdateModelConfigUseCase(modelConfigRepo, clk, bus)
	listModelsUC := modelconfigapp.NewListModelConfigsUseCase(modelConfigRepo)
	getModelUC := modelconfigapp.NewGetModelConfigUseCase(modelConfigRepo)

	// 8. 事件订阅（唯一注册点）
	if kbID := os.Getenv("DEFAULT_KB_ID"); kbID != "" {
		bus.Subscribe("ReportUploaded", knowledgeapp.NewReportUploadedHandler(
			addDocUC, reportRepo, knowledgedomain.NewKnowledgeBaseId(kbID)))
		logger.Info("subscribed ReportUploaded handler", "knowledgeBaseId", kbID)
	}
	bus.Subscribe("ModelConfigUpdated", func(ctx context.Context, event ddd.DomainEvent) error {
		ev, ok := event.(modelconfigdomain.ModelConfigUpdated)
		if !ok {
			return nil
		}
		mc, err := modelConfigRepo.FindByID(ctx, ev.ConfigID)
		if err != nil {
			return err
		}
		return llmFactory.OnConfigUpdated(ctx, mc)
	})
	bus.Subscribe("AgentPublished", func(ctx context.Context, event ddd.DomainEvent) error {
		if ev, ok := event.(agentdomain.AgentPublished); ok {
			logger.Info("agent published", "agentId", ev.AgentID.String(), "version", ev.Version)
		}
		return nil
	})

	// 9. 种子数据（首次启动可用的默认模型配置与内置工具）
	seed(ctx, logger, cfg, modelConfigRepo, toolRepo, ids, clk)

	// 10. 后台工作者
	worker := job.NewAnalysisWorker(runTaskUC, analysisRepo, logger)
	worker.Start(ctx, cfg.AnalysisWorkers)

	// 11. HTTP 接口
	router := httpiface.NewRouter(httpiface.Deps{
		Logger:              logger,
		Files:               files,
		UploadReport:        uploadUC,
		ParseReport:         parseUC,
		GetReport:           getReportUC,
		ListReports:         listReportsUC,
		CreateKnowledgeBase: createKBUC,
		AddDocument:         addDocUC,
		ReindexDocument:     reindexUC,
		SearchKnowledge:     searchUC,
		ListKnowledgeBases:  listKBUC,
		GetKnowledgeBase:    getKBUC,
		CreateDraftAgent:    createDraftUC,
		AddSkill:            addSkillUC,
		BindTool:            bindToolUC,
		PublishAgent:        publishUC,
		NewAgentVersion:     newVersionUC,
		RollbackAgent:       rollbackUC,
		ListAgents:          listAgentsUC,
		GetAgent:            getAgentUC,
		CreateConversation:  createConvUC,
		StartTurn:           startTurnUC,
		CompressContext:     compressUC,
		ListConversations:   listConvsUC,
		GetConversation:     getConvUC,
		Ask:                 askUC,
		AskStream:           askStreamUC,
		SubmitTask:          submitTaskUC,
		RunTask:             runTaskUC,
		GetResult:           getResultUC,
		ListTasks:           listTasksUC,
		RegisterTool:        registerToolUC,
		EnableTool:          enableToolUC,
		ListTools:           listToolsUC,
		GetTool:             getToolUC,
		CreateModelConfig:   createModelUC,
		UpdateModelConfig:   updateModelUC,
		ListModelConfigs:    listModelsUC,
		GetModelConfig:      getModelUC,
	})

	srv := &nethttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		// SSE 需要长连接，写超时必须为 0。
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			logger.Error("http server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// defaultModelClient 依据配置构造一个默认模型客户端（用于上下文摘要等辅助调用）。
func defaultModelClient(factory modelconfigdomain.ClientFactory, cfg *config.Config) (modelconfigdomain.ClientPort, string) {
	if cfg.DefaultModel == "" {
		return nil, ""
	}
	provider := modelconfigdomain.ProviderType(cfg.DefaultProvider)
	baseURL, apiKey := cfg.OpenAIBaseURL, cfg.OpenAIAPIKey
	switch provider {
	case modelconfigdomain.ProviderAnthropic:
		baseURL, apiKey = cfg.AnthropicBaseURL, cfg.AnthropicAPIKey
	case modelconfigdomain.ProviderOpenAI, modelconfigdomain.ProviderOpenAICompatible:
	default:
		provider = modelconfigdomain.ProviderOpenAI
	}
	if apiKey == "" {
		return nil, ""
	}
	tmp, err := modelconfigdomain.NewModelConfig(
		modelconfigdomain.NewModelConfigId("__bootstrap__"),
		"bootstrap", provider, baseURL, apiKey, cfg.DefaultModel,
		modelconfigdomain.DefaultParameters(), time.Now(),
	)
	if err != nil {
		return nil, ""
	}
	client, err := factory.Create(tmp)
	if err != nil {
		return nil, ""
	}
	return client, cfg.DefaultModel
}

// seed 首次启动时写入默认模型配置与内置工具（已存在则跳过）。
func seed(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
	modelConfigRepo modelconfigdomain.Repository,
	toolRepo tooldomain.Repository,
	ids idgen.Generator,
	clk clock.Clock,
) {
	if _, total, err := modelConfigRepo.List(ctx, 0, 1); err == nil && total == 0 && cfg.DefaultModel != "" {
		provider := modelconfigdomain.ProviderType(cfg.DefaultProvider)
		baseURL, apiKey := cfg.OpenAIBaseURL, cfg.OpenAIAPIKey
		if provider == modelconfigdomain.ProviderAnthropic {
			baseURL, apiKey = cfg.AnthropicBaseURL, cfg.AnthropicAPIKey
		}
		if apiKey != "" {
			mc, err := modelconfigdomain.NewModelConfig(
				modelconfigdomain.NewModelConfigId(ids.New()),
				"默认模型", provider, baseURL, apiKey, cfg.DefaultModel,
				modelconfigdomain.DefaultParameters(), clk.Now(),
			)
			if err == nil && modelConfigRepo.Save(ctx, mc) == nil {
				logger.Info("seeded default model config", "model", cfg.DefaultModel)
			}
		} else {
			logger.Warn("DEFAULT_MODEL set but no API key provided; skipping model config seed")
		}
	}

	if _, total, err := toolRepo.List(ctx, 0, 1); err == nil && total == 0 {
		for _, spec := range builtinToolSpecs() {
			t, err := tooldomain.NewTool(
				tooldomain.NewToolId(ids.New()),
				spec.name, spec.description, spec.schema,
				mustToolSource(), mustHandlerRef(spec.name), 30*time.Second, clk.Now(),
			)
			if err != nil {
				logger.Warn("seed tool failed", "tool", spec.name, "error", err)
				continue
			}
			if err := toolRepo.Save(ctx, t); err != nil {
				logger.Warn("persist seeded tool failed", "tool", spec.name, "error", err)
			}
		}
		logger.Info("seeded builtin tools")
	}
}

type toolSpec struct {
	name        string
	description string
	schema      tooldomain.JSONSchema
}

func builtinToolSpecs() []toolSpec {
	return []toolSpec{
		{
			name:        "calculator",
			description: "计算一个算术表达式，支持 + - * / % 与括号。",
			schema: tooldomain.ObjectSchema(map[string]tooldomain.JSONSchema{
				"expression": {Type: "string", Description: "需要计算的表达式，例如 (1+2)*3"},
			}, "expression"),
		},
		{
			name:        "http_request",
			description: "向指定 URL 发起 HTTP 请求并返回响应内容。",
			schema: tooldomain.ObjectSchema(map[string]tooldomain.JSONSchema{
				"url":     {Type: "string", Description: "目标 URL"},
				"method":  {Type: "string", Description: "HTTP 方法，默认 GET"},
				"headers": {Type: "object", Description: "请求头"},
				"body":    {Type: "string", Description: "请求体"},
			}, "url"),
		},
		{
			name:        "web_search",
			description: "联网搜索，返回与查询相关的网页摘要。",
			schema: tooldomain.ObjectSchema(map[string]tooldomain.JSONSchema{
				"query": {Type: "string", Description: "搜索关键词"},
			}, "query"),
		},
		{
			name:        "sql_query",
			description: "执行只读 SQL 查询（仅支持 SELECT）。",
			schema: tooldomain.ObjectSchema(map[string]tooldomain.JSONSchema{
				"sql": {Type: "string", Description: "只读 SQL 语句"},
			}, "sql"),
		},
	}
}

func mustToolSource() tooldomain.ToolSource {
	src, err := tooldomain.NewToolSource(tooldomain.SourceBuiltin, "", nil)
	if err != nil {
		panic(err)
	}
	return src
}

func mustHandlerRef(name string) tooldomain.HandlerRef {
	ref, err := tooldomain.NewHandlerRef(name)
	if err != nil {
		panic(err)
	}
	return ref
}
