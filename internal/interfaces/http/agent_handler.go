package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	agentapp "github.com/walry/ledger/internal/application/agent"
)

// AgentHandler 处理智能体配置相关请求。
type AgentHandler struct {
	createDraft *agentapp.CreateDraftUseCase
	addSkill    *agentapp.AddSkillUseCase
	bindTool    *agentapp.BindToolUseCase
	publish     *agentapp.PublishUseCase
	newVersion  *agentapp.NewVersionUseCase
	rollback    *agentapp.RollbackUseCase
	list        *agentapp.ListAgentsUseCase
	get         *agentapp.GetAgentUseCase
}

// NewAgentHandler 创建处理器。
func NewAgentHandler(
	createDraft *agentapp.CreateDraftUseCase,
	addSkill *agentapp.AddSkillUseCase,
	bindTool *agentapp.BindToolUseCase,
	publish *agentapp.PublishUseCase,
	newVersion *agentapp.NewVersionUseCase,
	rollback *agentapp.RollbackUseCase,
	list *agentapp.ListAgentsUseCase,
	get *agentapp.GetAgentUseCase,
) *AgentHandler {
	return &AgentHandler{
		createDraft: createDraft, addSkill: addSkill, bindTool: bindTool,
		publish: publish, newVersion: newVersion, rollback: rollback,
		list: list, get: get,
	}
}

// List 分页查询智能体。
func (h *AgentHandler) List(c *gin.Context) {
	res, err := h.list.Execute(c.Request.Context(), agentapp.ListAgentsCommand{
		Offset: queryInt(c, "offset", 0),
		Limit:  queryInt(c, "limit", 20),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

// Get 查询智能体详情。
func (h *AgentHandler) Get(c *gin.Context) {
	res, err := h.get.Execute(c.Request.Context(), agentapp.GetAgentCommand{AgentID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

type createAgentRequest struct {
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	SystemPrompt     string   `json:"systemPrompt" binding:"required"`
	ModelConfigID    string   `json:"modelConfigId" binding:"required"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
	MaxSteps         int      `json:"maxSteps"`
	RetrievalTopK    int      `json:"retrievalTopK"`
}

// CreateDraft 创建草稿智能体。
func (h *AgentHandler) CreateDraft(c *gin.Context) {
	var req createAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.createDraft.Execute(c.Request.Context(), agentapp.CreateDraftCommand{
		Name:             req.Name,
		Description:      req.Description,
		SystemPrompt:     req.SystemPrompt,
		ModelConfigID:    req.ModelConfigID,
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		MaxSteps:         req.MaxSteps,
		RetrievalTopK:    req.RetrievalTopK,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"agentId": res.AgentID, "version": res.Version})
}

type addSkillRequest struct {
	SkillID string `json:"skillId" binding:"required"`
	Version int    `json:"version"`
}

// AddSkill 绑定技能。
func (h *AgentHandler) AddSkill(c *gin.Context) {
	var req addSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.addSkill.Execute(c.Request.Context(), agentapp.AddSkillCommand{
		AgentID: c.Param("id"), SkillID: req.SkillID, Version: req.Version,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"agentId": res.AgentID})
}

type bindToolRequest struct {
	ToolID  string            `json:"toolId" binding:"required"`
	Version int               `json:"version"`
	Config  map[string]string `json:"config"`
}

// BindTool 绑定工具。
func (h *AgentHandler) BindTool(c *gin.Context) {
	var req bindToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.bindTool.Execute(c.Request.Context(), agentapp.BindToolCommand{
		AgentID: c.Param("id"), ToolID: req.ToolID, Version: req.Version, Config: req.Config,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"agentId": res.AgentID})
}

// Publish 发布智能体。
func (h *AgentHandler) Publish(c *gin.Context) {
	res, err := h.publish.Execute(c.Request.Context(), agentapp.PublishCommand{AgentID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"agentId": res.AgentID, "version": res.Version})
}

// NewVersion 基于当前版本派生新草稿版本。
func (h *AgentHandler) NewVersion(c *gin.Context) {
	res, err := h.newVersion.Execute(c.Request.Context(), agentapp.NewVersionCommand{AgentID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"agentId": res.AgentID, "version": res.Version, "parentVersion": res.ParentVersion})
}

// Rollback 以历史版本为蓝本创建新草稿版本。
func (h *AgentHandler) Rollback(c *gin.Context) {
	res, err := h.rollback.Execute(c.Request.Context(), agentapp.RollbackCommand{AgentID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"agentId": res.AgentID, "version": res.Version, "parentVersion": res.ParentVersion})
}
