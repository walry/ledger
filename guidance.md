# DDD 智能体应用建模最佳实践指南

## 目录
* 第一部分：核心原则

* 第二部分：限界上下文划分

* 第三部分：目录结构

* 第四部分：各层建模规范

* 第五部分：运行时编排规范

* 第六部分：装配层规范

* 第七部分：生成代码检查清单

* 第八部分：一句话心法

## 第一部分：核心原则（必须遵守）

### 原则 1：分层与依赖方向

```text
interfaces (HTTP/CLI) → application (用例) → domain (领域)
                                                ▲
                                                │ 实现端口
                                          infrastructure (技术实现)
```

* domain/ **绝不** import infrastructure/、net/http、第三方 SDK

* 接口定义在使用方（通常是领域层），实现在基础设施层

* 依赖倒置：领域层定义端口，基础设施层实现
  
### 原则 2：聚合边界

* 聚合是一致性边界，一次事务只修改一个聚合

* 跨聚合只能通过 ID / Ref（值对象） 引用，禁止对象引用

* 需要跨聚合协作时：领域服务（读多写一） 或 领域事件（最终一致）

* 如果两个聚合总是一起改、强一致，它们本就是一个聚合

### 原则 3：配置态 vs 运行态分离

* 配置态（Agent、Skill、Tool、ModelConfig）：低频变更、长期存在、版本化

* 运行态（Conversation、AnalysisTask）：高频变更、一次会话、用完即弃

* 两者必须分属不同聚合，绝不互相嵌套

### 原则 4：不要过度建模

* Tool 只有几个内置、不可自定义 → 值对象即可，不必聚合根

* 聚合构造简单 → 直接用构造函数，不必工厂

* 工厂只在需要 ID 生成、复杂组装、多创建路径时引入

## 第二部分：限界上下文划分

```text
┌─────────────────────────────────────────────────────┐
│              智能财务报告分析应用                      │
├─────────────────────────────────────────────────────┤
│  report       财务报告（业务对象：内容）              │
│  knowledge    知识库（技术能力：索引）                │
│  agent        智能体（配置态）                        │
│  conversation 会话（运行态）                          │
│  tool         工具（能力）                            │
│  modelconfig  模型配置（连接）                        │
│  runtime      运行时（跨上下文领域服务编排）           │
└─────────────────────────────────────────────────────┘
```

### 各上下文职责与聚合

上下文	职责	聚合根	值对象/实体
report	报告上传、解析、结构化管理	FinancialReport	Section、Table、ReportId
knowledge	知识库、文档、切片、向量	KnowledgeBase、Document	Chunk、Embedding、DocumentRef
agent	Agent 配置、版本、发布	Agent	SystemPrompt、AgentContext、SkillBinding、ToolBinding、ModelConfigRef
conversation	对话、轮次、上下文压缩	Conversation	Turn、Message、ConversationSummary
tool	工具定义、执行器注册	Tool	HandlerRef、JSONSchema
modelconfig	模型连接配置	ModelConfig	ModelParameters、ProviderType
runtime	RAG Pipeline、Agent 执行器	无（领域服务）	—

### 关键区分：report vs knowledge

维度	FinancialReport	KnowledgeBase
性质	业务对象（内容）	技术能力（索引）
生命周期	上传→解析→归档	索引→重建
关系	一份报告可索引到多个知识库	一个知识库可含多种文档
上传归属	上传是 report 的用例	索引是 knowledge 的用例
协作	发 ReportUploadedEvent	订阅事件，创建 Document

## 第三部分：目录结构（可直接用于生成代码）

```text
financial-report-analyzer/
├── cmd/api/main.go                      # 装配入口
├── internal/
│   ├── domain/                          # 领域层
│   │   ├── report/
│   │   │   ├── report.go                # 聚合根
│   │   │   ├── section.go               # 实体
│   │   │   ├── table.go                 # 值对象
│   │   │   ├── id.go
│   │   │   ├── status.go
│   │   │   ├── events.go
│   │   │   ├── repository.go            # 仓储接口
│   │   │   ├── factory.go               # 聚合工厂
│   │   │   └── service/
│   │   │       └── parser.go            # 端口：Parser
│   │   ├── knowledge/
│   │   │   ├── knowledge_base.go        # 聚合根 1
│   │   │   ├── document.go              # 聚合根 2
│   │   │   ├── chunk.go                 # Document 内实体
│   │   │   ├── embedding.go             # 值对象
│   │   │   ├── ref.go                   # DocumentRef 等
│   │   │   ├── id.go
│   │   │   ├── events.go
│   │   │   ├── repository.go            # 两个聚合的仓储接口
│   │   │   └── service/
│   │   │       ├── chunker.go           # 端口：Chunker
│   │   │       └── retriever.go         # 端口：Retriever
│   │   ├── agent/
│   │   │   ├── agent.go                 # 聚合根
│   │   │   ├── system_prompt.go         # 值对象
│   │   │   ├── agent_context.go         # 配置态上下文（值对象）
│   │   │   ├── binding.go               # SkillBinding / ToolBinding
│   │   │   ├── id.go
│   │   │   ├── status.go
│   │   │   ├── events.go
│   │   │   ├── repository.go
│   │   │   └── factory.go
│   │   ├── conversation/
│   │   │   ├── conversation.go          # 聚合根
│   │   │   ├── turn.go                  # 实体（轮级 Agent 绑定）
│   │   │   ├── message.go               # 值对象
│   │   │   ├── summary.go               # 值对象（压缩摘要）
│   │   │   ├── id.go
│   │   │   ├── repository.go
│   │   │   └── service/
│   │   │       └── compressor.go        # 领域服务：上下文压缩
│   │   ├── tool/
│   │   │   ├── tool.go                  # 聚合根
│   │   │   ├── handler.go               # 端口：Handler
│   │   │   ├── factory.go               # 端口：HandlerFactory
│   │   │   ├── schema.go
│   │   │   ├── source.go                # ToolSource 值对象
│   │   │   ├── id.go
│   │   │   └── repository.go
│   │   ├── modelconfig/
│   │   │   ├── model_config.go          # 聚合根
│   │   │   ├── parameters.go
│   │   │   ├── port.go                  # 端口：ClientPort / ClientFactory
│   │   │   ├── id.go
│   │   │   └── repository.go
│   │   └── runtime/                     # 跨上下文领域服务
│   │       ├── rag_pipeline.go
│   │       ├── agent_executor.go        # ReAct 循环
│   │       ├── qa_assistant.go
│   │       └── digital_employee.go
│   │
│   ├── application/                     # 应用层（按用例分包）
│   │   ├── report/
│   │   │   ├── upload.go                # UploadReportUseCase
│   │   │   ├── parse.go
│   │   │   ├── get.go
│   │   │   ├── list.go
│   │   │   └── dto.go
│   │   ├── knowledge/
│   │   │   ├── create_kb.go
│   │   │   ├── add_document.go
│   │   │   ├── reindex.go
│   │   │   ├── search.go
│   │   │   └── dto.go
│   │   ├── agent/
│   │   │   ├── create_draft.go
│   │   │   ├── add_skill.go
│   │   │   ├── bind_tool.go
│   │   │   ├── publish.go
│   │   │   ├── new_version.go
│   │   │   ├── rollback.go
│   │   │   └── dto.go
│   │   ├── conversation/
│   │   │   ├── create.go
│   │   │   ├── start_turn.go
│   │   │   ├── compress_context.go
│   │   │   └── dto.go
│   │   ├── qa/                          # ★ 问答助手业务用例
│   │   │   ├── ask.go
│   │   │   ├── ask_stream.go
│   │   │   └── dto.go
│   │   ├── analysis/                    # ★ 数字员工业务用例
│   │   │   ├── submit_task.go
│   │   │   ├── run_task.go
│   │   │   ├── get_result.go
│   │   │   └── dto.go
│   │   ├── tool/
│   │   │   ├── register.go
│   │   │   ├── enable.go
│   │   │   └── dto.go
│   │   └── modelconfig/
│   │       ├── create.go
│   │       ├── update.go
│   │       └── dto.go
│   │
│   ├── infrastructure/                  # 基础设施层
│   │   ├── persistence/
│   │   │   ├── report/
│   │   │   ├── knowledge/
│   │   │   ├── agent/
│   │   │   ├── conversation/
│   │   │   ├── tool/
│   │   │   └── modelconfig/
│   │   ├── vectorstore/
│   │   │   ├── milvus.go
│   │   │   └── retriever.go             # 实现 knowledge.Retriever
│   │   ├── llm/
│   │   │   ├── openai.go
│   │   │   ├── anthropic.go
│   │   │   ├── factory.go               # 实现 modelconfig.ClientFactory
│   │   │   └── streaming.go
│   │   ├── embedding/
│   │   │   └── openai_embedding.go
│   │   ├── parser/
│   │   │   ├── pdf_parser.go            # 实现 report.Parser
│   │   │   ├── excel_parser.go
│   │   │   └── word_parser.go
│   │   ├── toolhandler/
│   │   │   ├── registry.go              # 实现 tool.HandlerFactory
│   │   │   ├── builtin_web_search.go
│   │   │   ├── builtin_calculator.go
│   │   │   ├── builtin_http.go
│   │   │   └── builtin_sql_query.go
│   │   ├── eventbus/
│   │   │   └── in_memory.go
│   │   └── idgen/
│   │       └── uuid.go
│   │
│   └── interfaces/                      # 接口层
│       ├── http/
│       │   ├── report_handler.go
│       │   ├── knowledge_handler.go
│       │   ├── agent_handler.go
│       │   ├── qa_handler.go            # SSE 流式
│       │   ├── analysis_handler.go
│       │   ├── middleware/
│       │   └── router.go
│       ├── cli/
│       └── job/
│           └── analysis_worker.go
├── pkg/
└── configs/
```

## 第四部分：各层建模规范

### 4.1 聚合根规范

```go
// 结构：私有字段 + 只读访问器 + 领域行为 + 事件累积
type Agent struct {
	id           AgentId
	name         string
	status       AgentStatus
	systemPrompt SystemPrompt
	modelRef     ModelConfigRef
	skills       []SkillBinding
	tools        []ToolBinding
	context      AgentContext
	version      int
	events       []DomainEvent
}

// 构造函数：守护不变量
func NewAgent(id AgentId, name string, sp SystemPrompt, ref ModelConfigRef) (*Agent, error)

// 只读访问器
func (a *Agent) ID() AgentId
func (a *Agent) Status() AgentStatus

// 领域行为：所有修改都走方法
func (a *Agent) AddSkill(ref SkillBinding) error
func (a *Agent) Publish() error
func (a *Agent) NewVersion(newID AgentId) (*Agent, error)

// 事件管理
func (a *Agent) PullEvents() []DomainEvent
```

**铁律：**

* 字段全私有，方法改状态

* 构造时校验不变量

* 状态流转有规则（草稿→发布→归档，发布后不可改）

* 事件在状态变更时记录，由应用层 Pull 并发布

### 4.2 跨聚合引用规范

```go
// ✅ 正确：ID 引用 / Ref 值对象
type Conversation struct {
	agentID  agent.AgentId      // 纯 ID
	agentVer int                // 版本锁定
}

type Agent struct {
	skills []SkillBinding       // 含 ID + Version + 配置
	tools  []ToolBinding
}

type SkillBinding struct {
	SkillID skill.SkillId
	Version int
	Enabled bool
}

// ❌ 禁止：对象引用
type Conversation struct {
	agent *agent.Agent          // 禁止
}
```

**命名规范:**

* 只关联 ID：XxxId

* 需版本/配置：XxxBinding（推荐）或 XxxRef

* 历史快照：XxxSnapshot

### 4.3 端口与适配器规范

```go
// 领域层：定义端口（接口）
// domain/modelconfig/port.go
type ClientPort interface {
	Invoke(ctx context.Context, req Request) (Response, error)
	Stream(ctx context.Context, req Request) (<-chan Chunk, error)
}

type ClientFactory interface {
	CreateFor(config *ModelConfig) (ClientPort, error)
}

// 基础设施层：实现端口
// infrastructure/llm/factory.go
type DefaultClientFactory struct { ... }
func (f *DefaultClientFactory) CreateFor(cfg *modelconfig.ModelConfig) (modelconfig.ClientPort, error)
```

**规则：**

* 接口在使用方定义（通常是领域层）

* 产出领域对象的工厂（AgentFactory）实现在领域层

* 产出技术对象的工厂（ClientFactory）实现在基础设施层

* 端口方法签名用领域语言，不含技术细节

### 4.4 领域事件与 EventBus 规范

```go
// 领域层定义
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type EventBus interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}

// 基础设施层实现
type InMemoryBus struct {
	handlers map[string][]Handler
}
func (b *InMemoryBus) Subscribe(eventName string, h Handler)
func (b *InMemoryBus) Publish(ctx context.Context, events ...DomainEvent) error

// 装配层注册（唯一知道谁订阅谁的地方）
bus.Subscribe("ModelConfigUpdated", factory.OnModelConfigUpdated)
bus.Subscribe("ReportUploaded", addDocumentUC.OnReportUploaded)
bus.Subscribe("AgentPublished", agentCache.OnAgentPublished)
```

**规则：**

* 领域层只认识 DomainEvent 和 EventBus 接口

* 订阅关系在装配层注册

* Handler 必须幂等（事件可能重复投递）

* 生产环境用 Outbox 保证"保存+发事件"原子性

### 4.5 应用层用例规范

```go
// 一个用例一个文件
// application/agent/publish.go
package agent

// 输入 DTO
type PublishCommand struct {
	AgentID string
}

// 输出 DTO
type PublishResult struct {
	AgentID string
	Version int
}

// 用例
type PublishUseCase struct {
	agentRepo agent.Repository
	eventBus  agent.EventBus
}

func NewPublishUseCase(r agent.Repository, bus agent.EventBus) *PublishUseCase

func (uc *PublishUseCase) Execute(ctx context.Context, cmd PublishCommand) (PublishResult, error) {
	// 1. 加载聚合
	// 2. 跨聚合校验（应用层职责）
	// 3. 调领域行为
	// 4. 保存
	// 5. 发布事件
}
```

**规则：**

* 按用例分包，不按技术类型（不建全局 commands/queries/dtos）

* DTO 跟随用例

* 应用层只编排，不写业务逻辑，不直接调基础设施

* 跨聚合校验在应用层或领域服务

### 4.6 业务用例 vs 领域上下文

两个智能体（问答助手、数字员工）的差异体现在应用层：

```text
application/qa/ask.go               # 问答助手：RAG + 单轮/流式
application/analysis/submit_task.go  # 数字员工：多步 + 工具 + 可能异步
```

它们不拥有聚合，而是编排底层聚合，共享 runtime/ 领域服务。

### 4.7 上下文压缩规范

```go
// 领域服务
type CompressionStrategy interface {
	Compress(ctx context.Context, turns []conversation.Turn) (CompressedContext, error)
}

type ContextCompressor struct {
	strategy CompressionStrategy
}

// 策略：
// - TruncationStrategy：截断（最简单）
// - SummarizingStrategy：LLM 摘要
// - SlidingWindowSummaryStrategy：滑动窗口 + 滚动摘要（推荐）
```
**规则：**

* 摘要属于 Conversation 聚合（若"窗口+摘要"）

* 压缩调 LLM 通过 ClientPort 端口

* 压缩 Prompt 要求逐字保留 UUID、路径、URL 等标识符

* Tool 调用对不能被切散

## 第五部分：运行时编排规范

### 5.1 RAG Pipeline

```go
// domain/runtime/rag_pipeline.go
type RAGPipeline struct {
	retriever knowledge.Retriever
	reranker  knowledge.Reranker  // 可选
}

func (p *RAGPipeline) Retrieve(ctx context.Context, query string, kbIDs []string) ([]knowledge.Chunk, error)
```

### 5.2 Agent 执行器（ReAct 循环）

```go
// domain/runtime/agent_executor.go
type AgentExecutor struct {
	agentRepo    agent.Repository
	convRepo     conversation.Repository
	toolRepo     tool.Repository
	toolFactory  tool.HandlerFactory
	modelFactory modelconfig.ClientFactory
	retriever    knowledge.Retriever
	compressor   *service.ContextCompressor
}

func (e *AgentExecutor) Run(ctx context.Context, ...) error {
	for step := 0; step < maxSteps; step++ {
		// 1. 组装 prompt（system + skills + tools schema + history + RAG）
		// 2. 调模型
		// 3. 解析 tool_calls
		// 4. 无 tool_calls → 完成
		// 5. 执行工具，回填观察，继续循环
	}
}
```

### 5.3 SSE 流式响应

```go
// 领域层端口扩展
type ClientPort interface {
	Invoke(ctx, req) (Response, error)
	Stream(ctx, req) (<-chan Chunk, error)
}

// Gin handler
func (h *QAHandler) AskStream(c *gin.Context) {
	w := sse.NewWriter(c)  // 设置 text/event-stream + Flush
	chunks, _ := h.askUC.ExecuteStream(ctx, cmd)
	for chunk := range chunks {
		w.Send(sse.Event{Event: "message", Data: map[string]string{"delta": chunk.Delta}})
	}
	w.Send(sse.Event{Event: "done", Data: "[DONE]"})
}
```

**要点：**

* Content-Type: text/event-stream

* X-Accel-Buffering: no（禁用 Nginx 缓冲）

* 每次写完 Flush()

* 心跳注释保活

* WriteTimeout: 0

* 检测客户端断开（ctx.Done()）

* 流结束把完整回答落回 Conversation

## 第六部分：装配层规范

```go
// cmd/api/main.go
func main() {
	// 1. 基础设施
	db := ...
	reportRepo := persistence.NewReportRepo(db)
	kbRepo := persistence.NewKnowledgeRepo(db)
	retriever := vectorstore.NewMilvusRetriever(...)
	parser := parser.NewPDFParser()
	modelFactory := llm.NewClientFactory()
	toolRegistry := toolhandler.NewRegistry()
	idGen := idgen.NewUUIDGenerator()
	bus := eventbus.NewInMemoryBus()

	// 2. 领域服务
	compressor := service.NewContextCompressor(service.NewSlidingWindowSummaryStrategy(modelFactory))
	executor := runtime.NewAgentExecutor(agentRepo, convRepo, toolRepo, toolRegistry, modelFactory, retriever, compressor)
	ragPipeline := runtime.NewRAGPipeline(retriever)

	// 3. 事件订阅（唯一注册点）
	bus.Subscribe("ReportUploaded", addDocUC.OnReportUploaded)
	bus.Subscribe("ModelConfigUpdated", modelFactory.OnConfigUpdated)
	bus.Subscribe("AgentPublished", agentCache.OnAgentPublished)

	// 4. 应用用例
	askUC := qa.NewAskUseCase(agentRepo, convRepo, kbRepo, modelFactory, retriever, compressor)
	submitUC := analysis.NewSubmitTaskUseCase(agentRepo, reportRepo, taskRepo, toolRegistry, modelFactory, executor)
	uploadUC := report.NewUploadReportUseCase(reportRepo, parser, bus)

	// 5. HTTP
	router := http.NewRouter(askUC, submitUC, uploadUC, ...)
}
```

**规则：**

* 装配层是唯一知道"谁实现谁、谁订阅谁"的地方

* 领域层、应用层不感知基础设施

* 通过构造函数注入依赖

## 第七部分：生成代码时的检查清单

生成代码时，逐条自查：

### 领域层
□ 没有 import 基础设施、net/http、第三方 SDK
□ 每个聚合根字段私有，只有访问器和领域行为
□ 跨聚合只用 ID / Binding / Ref 引用，无对象引用
□ 每个聚合有独立仓储接口
□ 状态流转有不变量守护
□ 领域事件在状态变更时记录
□ 端口（Parser、Retriever、Chunker、Handler、ClientPort）定义在领域层
### 应用层
□ 按用例分包，不按技术类型
□ 一个用例一个文件，DTO 跟随
□ 只编排，不写业务逻辑
□ 跨聚合校验在这里或领域服务
□ 一次事务只改一个聚合
### 基础设施层
□ 实现领域层端口
□ 仓储实现领域层接口
□ EventBus、Factory、Handler 实现在这里
### 装配层
□ 所有订阅关系在这里注册
□ 所有依赖在这里注入
□ 领域层不知道基础设施的存在
### 运行时
□ RAG、ReAct、压缩是领域服务
□ 调 LLM 通过 ClientPort 端口
□ SSE 用 text/event-stream + Flush + 心跳

## 第八部分：一句话心法
领域层按子域分包，应用层按用例分包，业务角色（问答/分析）体现在应用层而非领域层；聚合是强一致边界，跨聚合只引用 ID，协作靠领域服务或事件；配置态与运行态严格分离；端口在领域层，适配器在基础设施层；装配层是唯一知道"谁实现谁、谁订阅谁"的地方；不过度建模，简单场景用值对象、构造函数即可。