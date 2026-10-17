// 与后端 HTTP 契约对应的类型定义。
//
// 注意大小写约定：
// - 由 struct DTO 序列化的「查询类」响应为 PascalCase（如 ReportID / Total）。
// - 由 gin.H 构造的「创建/操作类」响应为 camelCase（如 reportId / chunkCount）。
// - 请求体一律 camelCase。
// 详见后端 internal/interfaces/http 与各 application/<ctx>/dto.go。

/** 统一成功响应封装 */
export interface ApiEnvelope<T> {
  data: T
}

/** 统一错误响应封装 */
export interface ApiError {
  error: string
}

/** 分页请求通用参数 */
export interface PageQuery {
  offset?: number
  limit?: number
}

// ---------------------------------------------------------------------------
// 报告 report
// ---------------------------------------------------------------------------

export interface ReportSummary {
  ReportID: string
  Name: string
  Status: string
  SectionCount: number
}

export interface ListReportsResult {
  Reports: ReportSummary[]
  Total: number
}

export interface ReportSection {
  ID: string
  Type: string
  Title: string
  Level: number
  Order: number
  Content: string
}

export interface ReportDetail {
  ReportID: string
  Name: string
  Status: string
  ContentType: string
  Filename: string
  StoragePath: string
  FailureReason: string
  SectionCount: number
  Sections: ReportSection[]
}

/** 上传响应（camelCase） */
export interface UploadReportResult {
  reportId: string
}

/** 解析响应（camelCase） */
export interface ParseReportResult {
  reportId: string
  sectionCount: number
}

// ---------------------------------------------------------------------------
// 知识库 knowledge
// ---------------------------------------------------------------------------

export interface KnowledgeBaseSummary {
  KnowledgeBaseID: string
  Name: string
  Description: string
  EmbeddingModel: string
  Dimension: number
  DocumentCount: number
  CreatedAt: string
}

export interface ListKnowledgeBasesResult {
  KnowledgeBases: KnowledgeBaseSummary[]
  Total: number
}

export interface KnowledgeDocument {
  DocumentID: string
  Title: string
  SourceKind: string
  SourceRefID: string
  Status: string
  FailureReason: string
  ChunkCount: number
  CreatedAt: string
}

export interface KnowledgeBaseDetail {
  KnowledgeBaseID: string
  Name: string
  Description: string
  EmbeddingModel: string
  Dimension: number
  DocumentCount: number
  CreatedAt: string
  UpdatedAt: string
  Documents: KnowledgeDocument[]
}

export interface SearchChunk {
  KnowledgeBaseID: string
  DocumentID: string
  ChunkID: string
  Content: string
  Score: number
  Metadata: Record<string, string>
}

export interface SearchResult {
  Chunks: SearchChunk[]
}

export interface CreateKnowledgeBasePayload {
  name: string
  description?: string
  embeddingModel?: string
  dimension?: number
}

export interface AddDocumentPayload {
  title?: string
  sourceKind?: string
  sourceRefId?: string
  text: string
}

export interface SearchPayload {
  query: string
  knowledgeBaseIds?: string[]
  topK?: number
}

/** 创建知识库响应（camelCase） */
export interface CreateKnowledgeBaseResult {
  knowledgeBaseId: string
}

/** 新增文档 / 重建索引响应（camelCase） */
export interface DocumentIndexResult {
  documentId: string
  chunkCount: number
}

// ---------------------------------------------------------------------------
// 智能体 agent
// ---------------------------------------------------------------------------

export interface AgentSummary {
  AgentID: string
  Name: string
  Description: string
  Status: string
  Version: number
  ModelConfigID: string
  Model: string
  Provider: string
}

export interface ListAgentsResult {
  Agents: AgentSummary[]
  Total: number
}

export interface SkillBinding {
  SkillID: string
  Version: number
  Enabled: boolean
}

export interface ToolBinding {
  ToolID: string
  Version: number
  Enabled: boolean
  Config: Record<string, string>
}

export interface AgentDetail {
  AgentID: string
  Name: string
  Description: string
  Status: string
  SystemPrompt: string
  ModelConfigID: string
  Model: string
  Provider: string
  KnowledgeBaseIDs: string[]
  MaxSteps: number
  RetrievalTopK: number
  Version: number
  ParentVersion: number
  CreatedAt: string
  UpdatedAt: string
  PublishedAt: string | null
  Skills: SkillBinding[]
  Tools: ToolBinding[]
}

export interface CreateAgentPayload {
  name: string
  description?: string
  systemPrompt: string
  modelConfigId: string
  knowledgeBaseIds?: string[]
  maxSteps?: number
  retrievalTopK?: number
}

export interface AddSkillPayload {
  skillId: string
  version?: number
}

export interface BindToolPayload {
  toolId: string
  version?: number
  config?: Record<string, string>
}

/** 创建草稿响应（camelCase） */
export interface AgentVersionResult {
  agentId: string
  version: number
}

/** 派生 / 回滚响应（camelCase） */
export interface AgentVersionChangeResult {
  agentId: string
  version: number
  parentVersion: number
}

/** 绑定技能 / 工具响应（camelCase） */
export interface AgentActionRef {
  agentId: string
}

// ---------------------------------------------------------------------------
// 会话 conversation
// ---------------------------------------------------------------------------

export interface ConversationSummary {
  ConversationID: string
  Title: string
  AgentID: string
  AgentVersion: number
  Status: string
  CreatedAt: string
}

export interface ListConversationsResult {
  Conversations: ConversationSummary[]
  Total: number
}

export interface ChatToolCall {
  ID: string
  Name: string
  Arguments: string
}

export interface ChatMessage {
  Role: string
  Content: string
  Name: string
  ToolCalls: ChatToolCall[]
  CreatedAt: string
}

export interface ConversationTurn {
  TurnID: string
  Index: number
  AgentVersion: number
  Status: string
  Failure: string
  CreatedAt: string
  Messages: ChatMessage[]
}

export interface ConversationDetail {
  ConversationID: string
  Title: string
  AgentID: string
  AgentVersion: number
  Status: string
  Summary: string
  CreatedAt: string
  UpdatedAt: string
  Turns: ConversationTurn[]
}

/** 创建会话响应（camelCase） */
export interface CreateConversationResult {
  conversationId: string
  agentId: string
  agentVersion: number
}

export interface CreateConversationPayload {
  agentId: string
  title?: string
}

/** 开启轮次响应（camelCase） */
export interface StartTurnResult {
  conversationId: string
  turnId: string
}

/** 压缩上下文响应（camelCase） */
export interface CompressContextResult {
  conversationId: string
  summary: string
  compressed: boolean
}

// ---------------------------------------------------------------------------
// 问答 qa
// ---------------------------------------------------------------------------

export interface AskResult {
  Answer: string
  TurnID: string
  Tips: string[]
}

export interface AskPayload {
  conversationId: string
  question: string
  knowledgeBaseIds?: string[]
}

export interface TokenUsage {
  promptTokens: number
  completionTokens: number
  totalTokens: number
}

// ---------------------------------------------------------------------------
// 分析任务 analysis
// ---------------------------------------------------------------------------

export interface AnalysisTaskSummary {
  TaskID: string
  Title: string
  Status: string
  ConversationID: string
  CreatedAt: string
}

export interface ListTasksResult {
  Tasks: AnalysisTaskSummary[]
  Total: number
}

export interface AnalysisResult {
  TaskID: string
  Title: string
  Status: string
  Result: string
  Failure: string
  ConversationID: string
  ReportIDs: string[]
}

export interface SubmitTaskPayload {
  title?: string
  question: string
  agentId: string
  reportIds?: string[]
}

/** 提交任务响应（camelCase） */
export interface SubmitTaskResult {
  taskId: string
}

/** 执行任务响应（camelCase） */
export interface RunTaskResult {
  taskId: string
  status: string
}

// ---------------------------------------------------------------------------
// 工具 tool
// ---------------------------------------------------------------------------

export interface JSONSchema {
  type?: string
  description?: string
  properties?: Record<string, JSONSchema>
  required?: string[]
  items?: JSONSchema
  enum?: string[]
}

export interface ToolSummary {
  ToolID: string
  Name: string
  Description: string
  Enabled: boolean
  SourceKind: string
  Version: number
}

export interface ListToolsResult {
  Tools: ToolSummary[]
  Total: number
}

export interface ToolDetail {
  ToolID: string
  Name: string
  Description: string
  Enabled: boolean
  Version: number
  SourceKind: string
  SourceURI: string
  Handler: string
  Schema: JSONSchema
  TimeoutSeconds: number
  CreatedAt: string
  UpdatedAt: string
}

export interface RegisterToolPayload {
  name: string
  description?: string
  handlerName: string
  schema: JSONSchema
  sourceKind?: string
  sourceUri?: string
  sourceConfig?: Record<string, string>
  timeoutSeconds?: number
}

/** 注册工具响应（camelCase） */
export interface RegisterToolResult {
  toolId: string
}

/** 启用/停用响应（camelCase） */
export interface EnableToolResult {
  toolId: string
  enabled: boolean
}

// ---------------------------------------------------------------------------
// 模型配置 modelconfig
// ---------------------------------------------------------------------------

export interface ModelParameters {
  Temperature: number
  TopP: number
  MaxTokens: number
  PresencePenalty: number
  FrequencyPenalty: number
  Stop: string[]
}

export interface ModelConfig {
  ConfigID: string
  Name: string
  Provider: string
  BaseURL: string
  Model: string
  Parameters: ModelParameters
  Enabled: boolean
  CreatedAt: string
  UpdatedAt: string
}

export interface ListModelConfigsResult {
  Configs: ModelConfig[]
  Total: number
}

export interface CreateModelConfigPayload {
  name: string
  provider: string
  baseUrl?: string
  apiKey?: string
  model: string
  parameters?: ModelParameters
}

export interface UpdateModelConfigPayload {
  baseUrl?: string
  apiKey?: string
  model: string
  parameters?: ModelParameters
}

/** 创建 / 更新模型配置响应（camelCase） */
export interface ModelConfigRef {
  configId: string
}
