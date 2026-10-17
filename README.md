# ledger · 智能财务报告分析应用

一个基于 **DDD（领域驱动设计）+ 端口适配器** 架构的 Go 后端服务，把「财务报告」变成可检索、可问答、可自动分析的知识资产。

应用提供三类业务能力：

- **报告管理**：上传、解析（PDF / Excel / Word / 纯文本）、结构化管理财务报告。
- **知识库**：文档切片、向量化（pgvector）、语义检索（RAG）。
- **智能体**：配置化的 Agent（系统提示词 + 技能 + 工具 + 模型 + 知识库），支持发布、版本、回滚。
  - 业务角色一 · **问答助手**：基于 RAG + ReAct 的单轮 / SSE 流式问答。
  - 业务角色二 · **数字员工**：面向报告的异步多步分析任务（后台 worker 消费队列）。

设计规范详见 [`guidance.md`](guidance.md)。

---

## 目录

- [架构总览](#架构总览)
- [限界上下文](#限界上下文)
- [目录结构](#目录结构)
- [快速开始](#快速开始)
- [Docker Compose 部署](#docker-compose-部署)
- [配置项](#配置项)
- [HTTP API](#http-api)
- [核心流程](#核心流程)
- [前端（frontend/）](#前端frontend)
- [开发](#开发)

---

## 架构总览

严格的分层依赖方向（依赖倒置，领域层定义端口、基础设施层实现）：

```text
interfaces (HTTP)  →  application (用例)  →  domain (领域)
                                                  ▲
                                                  │ 实现端口
                                         infrastructure (技术)
```

装配入口在 `cmd/api/main.go`，它是**唯一**知道「谁实现谁、谁订阅谁」的地方：

- 创建基础设施（数据库、LLM、向量库、解析器、事件总线……）
- 构造领域服务（RAG Pipeline、Agent 执行器、上下文压缩器、数字员工）
- 实例化应用用例并注册领域事件订阅
- 挂载 HTTP 路由

领域层不 import 基础设施、`net/http` 或任何第三方 SDK；跨聚合只通过 ID / Binding / Ref 引用，协作靠领域服务或领域事件。

---

## 限界上下文

| 上下文 | 职责 | 聚合根 | 关键值对象 / 实体 |
| --- | --- | --- | --- |
| `report` | 报告上传、解析、结构化管理 | `Report` | `Section`、`Table`、`ReportId` |
| `knowledge` | 知识库、文档、切片、向量 | `KnowledgeBase`、`Document` | `Chunk`、`Embedding`、`DocumentRef` |
| `agent` | Agent 配置、版本、发布 | `Agent` | `SystemPrompt`、`AgentContext`、`SkillBinding`、`ToolBinding`、`ModelConfigRef` |
| `conversation` | 对话、轮次、上下文压缩 | `Conversation` | `Turn`、`Message`、`ConversationSummary` |
| `analysis` | 数字员工分析任务 | `AnalysisTask` | 状态机（queued→running→succeeded/failed） |
| `tool` | 工具定义、执行器注册 | `Tool` | `HandlerRef`、`JSONSchema`、`ToolSource` |
| `modelconfig` | 模型连接配置 | `ModelConfig` | `ModelParameters`、`ProviderType` |
| `runtime` | RAG Pipeline、Agent 执行器、业务角色 | 无（领域服务） | — |

**配置态 vs 运行态**：`agent` / `tool` / `modelconfig` 是低频变更、版本化的配置态；`conversation` / `analysis` 是高频变更、用完即弃的运行态。两者分属不同聚合，绝不互相嵌套。

**关键区分 report vs knowledge**：`report` 是业务对象（内容），上传是 report 的用例；`knowledge` 是技术能力（索引），索引是 knowledge 的用例。报告上传后通过 `ReportUploaded` 领域事件，由 knowledge 上下文订阅并创建文档。

---

## 目录结构

```text
ledger/
├── cmd/api/main.go                 # 装配入口（唯一 DI / 订阅点）
├── configs/config.example.yaml     # 环境变量参考（应用不解析此文件）
├── internal/
│   ├── domain/                     # 领域层（按子域分包）
│   │   ├── report/                 # 报告聚合 + Parser 端口
│   │   ├── knowledge/              # 知识库/文档聚合 + Chunker/Retriever/Embedder 端口
│   │   ├── agent/                  # Agent 聚合 + 工厂
│   │   ├── conversation/           # 会话聚合 + service/ 上下文压缩
│   │   ├── analysis/               # 分析任务聚合
│   │   ├── tool/                   # 工具聚合 + Handler/HandlerFactory 端口
│   │   ├── modelconfig/            # 模型配置聚合 + ClientPort/ClientFactory 端口
│   │   ├── runtime/                # ★ 领域服务：RAG、ReAct 执行器、问答助手、数字员工
│   │   └── ddd/                    # DomainEvent / EventBus / Handler 基础抽象
│   ├── application/                # 应用层（按用例分包，一个用例一个文件）
│   │   ├── report/ knowledge/ agent/ conversation/
│   │   ├── qa/                     # ★ 问答助手用例（含流式）
│   │   ├── analysis/               # ★ 数字员工用例（提交/执行/查询）
│   │   ├── tool/ modelconfig/
│   ├── infrastructure/             # 基础设施层（端口实现）
│   │   ├── persistence/            # GORM 仓储（逐上下文）+ AutoMigrate
│   │   ├── vectorstore/            # pgvector 检索实现
│   │   ├── llm/                    # OpenAI / Anthropic 客户端 + 工厂 + 流式
│   │   ├── embedding/              # OpenAI 兼容 embedding
│   │   ├── parser/                 # PDF / Excel / Word / Text 解析
│   │   ├── chunker/                # 递归切片
│   │   ├── toolhandler/            # 内置工具：calculator / http_request / web_search / sql_query
│   │   ├── eventbus/               # 内存事件总线
│   │   ├── filestore/              # 本地文件存储
│   │   ├── idgen/ config/
│   └── interfaces/                 # 接口层
│       ├── http/                   # Gin 路由 + 各上下文 handler + SSE writer
│       └── job/                    # 后台 AnalysisWorker
├── pkg/
│   ├── clock/                      # 时钟抽象
│   └── idgen/                      # ID 生成抽象
└── go.mod
```

---

## 快速开始

### 前置依赖

- **Go 1.24+**
- **PostgreSQL**，并安装 **pgvector** 扩展（向量检索依赖 `CREATE EXTENSION vector`，需服务端已装扩展且当前用户有权限）
- 一个大模型接入（OpenAI 兼容接口或 Anthropic），以及一个 embedding 模型

> 未配置 LLM / embedding 时服务仍可启动，但 RAG 检索、问答、摘要会不可用；上下文压缩会退化为纯截断策略。

### 1. 准备数据库

```bash
createdb ledger
# 若 pgvector 尚未就绪（需管理员权限）：
psql -d ledger -c 'CREATE EXTENSION IF NOT EXISTS vector;'
```

### 2. 配置环境变量

复制参考文件并按需填写：

```bash
cp configs/config.example.yaml .env   # 内容仅供参考，按 shell/部署平台注入环境变量
```

最小可运行配置：

```bash
export DATABASE_DSN="postgres://user:pass@localhost:5432/ledger?sslmode=disable"
export OPENAI_API_KEY="sk-..."
export OPENAI_BASE_URL="https://api.openai.com/v1"
export DEFAULT_MODEL="gpt-4o-mini"          # 可选：用于摘要等辅助调用
export AUTO_MIGRATE=true
```

### 3. 运行

```bash
go run ./cmd/api
# 启动后监听 :8080，健康检查：curl localhost:8080/healthz
```

首次启动会自动建表（`AUTO_MIGRATE=true`），并**播种**：
- 一个默认模型配置（当设置了 `DEFAULT_MODEL` 且提供了 API Key）
- 四个内置工具：`calculator`、`http_request`、`web_search`、`sql_query`

### 4. 冒烟测试

```bash
# 上传报告
curl -F "file=@report.pdf" localhost:8080/api/reports

# 构造一个 Agent 草稿并发布（modelConfigId 用播种出的默认配置）
curl -X POST localhost:8080/api/agents -H 'Content-Type: application/json' -d '{
  "name":"财报助手","systemPrompt":"你是财报分析专家","modelConfigId":"<id>"
}'
```

---

## Docker Compose 部署

一键拉起 `db`（PostgreSQL + pgvector）、`api`（Go 后端）、`web`（Nginx + 前端静态资源）三个服务。
镜像约定：

| 服务 | 基础镜像 | 说明 |
| --- | --- | --- |
| 构建 Go | `golang:1.24-bookworm` | 多阶段构建，产出静态二进制 |
| 运行 API | `debian:12-slim` | 非 root 用户运行，含健康检查 |
| 构建前端 | `node:24-bookworm` | `npm ci` + `vite build` |
| 运行前端 | `nginx:stable-alpine` | 静态托管 + `/api` 反代 |
| 数据库 | `pgvector/pgvector:pg16` | 官方 PostgreSQL 16 + pgvector 扩展 |

### 1. 准备环境变量

```bash
cp .env.example .env
# 至少修改 POSTGRES_PASSWORD，并填写 OPENAI_API_KEY 等模型配置
```

### 2. 启动

```bash
docker compose up -d --build
docker compose ps          # 等待 db / api 变为 healthy
curl localhost/healthz     # 经 Nginx 反代到后端
```

- 前端入口：`http://localhost`（`WEB_PORT` 可改）
- 数据库仅绑定 `127.0.0.1:${POSTGRES_PORT}`，生产环境可移除该端口映射
- 上传文件持久化在 `uploads` 卷，数据库数据在 `db_data` 卷
- 首次启动 `AUTO_MIGRATE=true` 自动建表，并执行 `deploy/postgres/init/01-vector.sql` 创建 `vector` 扩展

### 3. 常用运维

```bash
docker compose logs -f api       # 查看后端日志
docker compose restart api       # 重启后端
docker compose down              # 停止并保留数据卷
docker compose down -v           # 停止并删除数据卷（谨慎）
```

### 4. 文件布局

```text
docker-compose.yml
.env.example
deploy/
├── docker/
│   ├── api.Dockerfile
│   └── web.Dockerfile
├── nginx/default.conf           # SPA 回退、/api 反代、SSE 关闭缓冲
└── postgres/init/01-vector.sql  # 初始化 pgvector 扩展
```

> 修改 `VECTOR_DIMENSION` 或 `EMBEDDING_DIMENSION` 后，已有 `knowledge_vectors` 表的向量列维度不会自动变更，需要重建该表（或重建数据卷）。

---

## 配置项

应用**完全通过环境变量**读取配置（`internal/infrastructure/config`），不解析外部配置文件。

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `DATABASE_DSN` | 空（**必填**） | PostgreSQL DSN |
| `AUTO_MIGRATE` | `true` | 启动时自动建表 |
| `OPENAI_API_KEY` | 空 | OpenAI 兼容 Key |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | OpenAI 兼容 Base URL |
| `ANTHROPIC_API_KEY` | 空 | Anthropic Key |
| `ANTHROPIC_BASE_URL` | `https://api.anthropic.com` | Anthropic Base URL |
| `EMBEDDING_MODEL` | `text-embedding-3-small` | Embedding 模型名 |
| `EMBEDDING_DIMENSION` | `1536` | Embedding 维度 |
| `VECTOR_DIMENSION` | `1536` | pgvector 列维度（须与 embedding 维度一致） |
| `DEFAULT_MODEL` | 空 | 默认模型（摘要等辅助调用 + 首次播种模型配置） |
| `DEFAULT_PROVIDER` | `openai` | 默认供应商：`openai` / `anthropic` / `openai_compatible` |
| `WEB_SEARCH_ENDPOINT` | 空 | `web_search` 工具的后端地址（空则工具不可用） |
| `WEB_SEARCH_API_KEY` | 空 | 网络搜索 Key |
| `SQL_DSN` | 空 | `sql_query` 工具的数据源（空则禁用，仅允许 SELECT） |
| `ANALYSIS_WORKERS` | `2` | 数字员工分析任务的并发 worker 数 |
| `UPLOAD_DIR` | `$TMPDIR/ledger-uploads` | 上传报告原始文件的本地存储目录 |
| `DEFAULT_KB_ID` | 空 | 若设置，则订阅 `ReportUploaded`，自动把新报告索引到该知识库 |

---

## HTTP API

所有响应统一封装：成功为 `{"data": ...}`，失败为 `{"error": "..."}`。路由前缀 `/api`。

### 报告

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/reports` | 上传报告（`multipart/form-data`：`file` 必填，`name` 可选） |
| `GET` | `/api/reports` | 分页列表（query：`status`、`offset`、`limit`） |
| `GET` | `/api/reports/:id` | 报告详情（含章节、存储路径等） |
| `POST` | `/api/reports/:id/parse` | 解析报告（从存储读取原文并结构化） |

### 知识库

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/knowledge-bases` | 分页列表（query：`offset`、`limit`） |
| `POST` | `/api/knowledge-bases` | 创建知识库 `{name, description?, embeddingModel?, dimension?}` |
| `GET` | `/api/knowledge-bases/:id` | 知识库详情（含文档列表） |
| `POST` | `/api/knowledge-bases/:id/documents` | 新增文档并索引 `{title?, sourceKind?, sourceRefId?, text}` |
| `POST` | `/api/documents/:id/reindex` | 重建文档索引 `{text?}` |
| `POST` | `/api/knowledge/search` | 语义检索 `{query, knowledgeBaseIds?, topK?}` |

### 智能体

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/agents` | 分页列表（query：`offset`、`limit`） |
| `POST` | `/api/agents` | 创建草稿 `{name, description?, systemPrompt, modelConfigId, knowledgeBaseIds?, maxSteps?, retrievalTopK?}` |
| `GET` | `/api/agents/:id` | 智能体详情（含技能/工具绑定） |
| `POST` | `/api/agents/:id/skills` | 绑定技能 `{skillId, version?}` |
| `POST` | `/api/agents/:id/tools` | 绑定工具 `{toolId, version?, config?}` |
| `POST` | `/api/agents/:id/publish` | 发布（发布后不可再修改） |
| `POST` | `/api/agents/:id/versions` | 派生新草稿版本 |
| `POST` | `/api/agents/:id/rollback` | 以历史版本为蓝本创建新草稿 |

### 会话

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/conversations` | 分页列表（query：`agentId?`、`offset`、`limit`） |
| `POST` | `/api/conversations` | 创建会话 `{agentId, title?}` |
| `GET` | `/api/conversations/:id` | 会话详情（含轮次与消息） |
| `POST` | `/api/conversations/:id/turns` | 记录一轮用户消息 `{message}` |
| `POST` | `/api/conversations/:id/compress` | 压缩会话上下文（滚动摘要） |

### 问答助手

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/qa/ask` | 非流式问答 `{conversationId, question, knowledgeBaseIds?}` |
| `POST` | `/api/qa/ask/stream` | 流式问答（SSE），body 同上 |
| `GET` | `/api/qa/ask/stream` | 流式问答（SSE），query：`conversationId`、`question` |

**SSE 事件流**（`text/event-stream`，含 15s 心跳注释保活）：

| event | data |
| --- | --- |
| `message` | `{"delta":"..."}` 增量文本 |
| `tool_call` | `{"id":"...","name":"..."}` |
| `tool_result` | `{"id":"...","name":"...","result":"..."}` |
| `done` | `{"answer":"...","usage":{...}}` |
| `error` | `{"error":"..."}` |

### 数字员工（分析任务）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/analysis/tasks` | 分页列表（query：`offset`、`limit`） |
| `POST` | `/api/analysis/tasks` | 提交任务（入队）`{title?, question, agentId, reportIds?}` |
| `POST` | `/api/analysis/tasks/:id/run` | 同步执行任务 |
| `GET` | `/api/analysis/tasks/:id` | 查询任务状态与结果 |

> 提交后的任务默认由后台 `AnalysisWorker` 异步消费（轮询间隔 2s，并发数见 `ANALYSIS_WORKERS`）。

### 工具 & 模型配置

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/tools` | 分页列表（query：`offset`、`limit`） |
| `POST` | `/api/tools` | 注册工具 `{name, description?, handlerName, schema, sourceKind?, sourceUri?, sourceConfig?, timeoutSeconds?}` |
| `GET` | `/api/tools/:id` | 工具详情（含 Schema） |
| `POST` | `/api/tools/:id/enable` | 启用/停用 `{enabled}` |
| `GET` | `/api/model-configs` | 分页列表（不返回 API Key） |
| `POST` | `/api/model-configs` | 创建模型配置 `{name, provider, baseUrl?, apiKey?, model, parameters?}` |
| `GET` | `/api/model-configs/:id` | 模型配置详情（不返回 API Key） |
| `PUT` | `/api/model-configs/:id` | 更新模型配置 `{model, baseUrl?, apiKey?, parameters?}` |

### 其他

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/healthz` | 健康检查 |

---

## 核心流程

### 报告上传 → 解析 → 索引

```text
POST /api/reports              保存原文件到本地存储，Report 聚合落库（状态 uploaded）
        │  （若设置 DEFAULT_KB_ID，装配层已订阅 ReportUploaded）
        ▼
ReportUploaded 事件 ──► knowledge 事件处理器：抽取报告全文 → AddDocument（切片 + 向量化 + pgvector 写入）
POST /api/reports/:id/parse    解析原文为结构化 Section/Table（状态 uploaded → parsed）
```

报告状态机：`uploaded → parsing → parsed → archived`，解析失败进入 `failed`（可重试）。

### 问答助手（RAG + ReAct）

应用层用例 `qa.AskUseCase` / `qa.AskStreamUseCase` 编排 `QAAssistant` 领域服务，后者复用 `AgentExecutor`：

```text
组装 prompt（system + skills + tools schema + history + RAG 检索）
  → 调模型 → 解析 tool_calls → 执行工具回填观察 → 循环（不超过 maxSteps）
  → 无 tool_calls 即完成，将完整回答落回 Conversation
```

流式模式下逐 token 通过 `ClientPort.Stream` 产出 `delta`，工具调用/结果作为独立 SSE 事件下发。

### 数字员工（异步分析）

```text
POST /api/analysis/tasks        AnalysisTask 落库（queued）
        ▼
AnalysisWorker（后台）            ClaimNextQueued → runTask.Execute
        ▼
DigitalEmployee 领域服务         读取任务 → 新建 Conversation → 拼接报告正文 → AgentExecutor.Run
        ▼
回写结果                          succeeded（result）/ failed（failure）
```

### 上下文压缩

`conversation/service.ContextCompressor` 支持多种策略：`Truncation`（截断）、`Summarizing`（LLM 摘要）、`SlidingWindowSummary`（滑动窗口 + 滚动摘要，**推荐**，启动时若配置了默认模型则自动启用）。压缩 prompt 要求逐字保留 UUID、路径、URL 等标识符，且不切断工具调用对。

---

## 前端（frontend/）

`frontend/` 是一个独立的 Vue 3 单页应用，覆盖全部模块（报告 / 知识库 / 智能体 / 会话 / 问答 / 分析 / 工具 / 模型配置），通过 `/api` 调用后端。

**技术栈**：Vue 3 + TypeScript + Vite + Pinia + Vue Router + Axios + Ant Design Vue 4。

```bash
cd frontend
npm install
npm run dev      # http://localhost:5173
npm run build    # vue-tsc 类型检查 + 产出 dist/
npm run preview  # 预览构建产物
```

**约定与要点**

- 开发时由 Vite 把 `/api` 代理到 `http://localhost:8080`（后端未配置 CORS，用代理规避跨域；可用 `VITE_PROXY_TARGET` 覆盖目标地址）。
- 请求基路径由 `VITE_API_BASE` 控制（默认 `/api`）。
- 后端响应大小写不一致：**查询类** 响应为 PascalCase（如 `ReportID`、`Total`），**创建/操作类** 响应与请求体为 camelCase（如 `reportId`、`chunkCount`）。类型定义见 `frontend/src/types/index.ts`。
- 问答使用 SSE 流式：因 `EventSource` 无法自定义/中止，改用 `fetch + ReadableStream` 手工解析（`frontend/src/api/sse.ts`）。
- 统一响应封装 `{"data": ...}` / `{"error": ...}` 在 `frontend/src/api/client.ts` 的拦截器中拆包与报错。

**目录结构**

```text
frontend/src/
├── api/          # client（axios+拦截器）、sse（流式解析）、各模块 API
├── components/   # PageHeader、StatusTag
├── layouts/      # MainLayout（侧边菜单 + 顶栏）
├── router/       # 路由
├── stores/       # Pinia（智能体、模型配置下拉缓存）
├── types/        # 与后端契约对应的 TS 类型
├── utils/        # 格式化工具
└── views/        # 各模块页面（reports/knowledge/agents/conversations/qa/analysis/tools/modelconfigs）
```

---

## 开发

```bash
go build ./...        # 编译
go test ./...         # 运行测试（当前覆盖 domain/agent、domain/report、domain/runtime）
go vet ./...          # 静态检查
go run ./cmd/api      # 本地启动
```

前端开发见 [前端（frontend/）](#前端frontend)。

代码规范与建模约束（聚合边界、端口与适配器、事件总线、SSE、检查清单等）见 [`guidance.md`](guidance.md)。
