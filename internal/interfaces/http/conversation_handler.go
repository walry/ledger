package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	conversationapp "github.com/walry/ledger/internal/application/conversation"
)

// ConversationHandler 处理会话相关请求。
type ConversationHandler struct {
	create   *conversationapp.CreateConversationUseCase
	start    *conversationapp.StartTurnUseCase
	compress *conversationapp.CompressContextUseCase
	list     *conversationapp.ListConversationsUseCase
	get      *conversationapp.GetConversationUseCase
}

// NewConversationHandler 创建处理器。
func NewConversationHandler(
	create *conversationapp.CreateConversationUseCase,
	start *conversationapp.StartTurnUseCase,
	compress *conversationapp.CompressContextUseCase,
	list *conversationapp.ListConversationsUseCase,
	get *conversationapp.GetConversationUseCase,
) *ConversationHandler {
	return &ConversationHandler{create: create, start: start, compress: compress, list: list, get: get}
}

// List 分页查询会话；可用 agentId 过滤。
func (h *ConversationHandler) List(c *gin.Context) {
	res, err := h.list.Execute(c.Request.Context(), conversationapp.ListConversationsCommand{
		AgentID: c.Query("agentId"),
		Offset:  queryInt(c, "offset", 0),
		Limit:   queryInt(c, "limit", 20),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

// Get 查询会话详情（含轮次与消息）。
func (h *ConversationHandler) Get(c *gin.Context) {
	res, err := h.get.Execute(c.Request.Context(), conversationapp.GetConversationCommand{ConversationID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

type createConversationRequest struct {
	AgentID string `json:"agentId" binding:"required"`
	Title   string `json:"title"`
}

// Create 创建会话。
func (h *ConversationHandler) Create(c *gin.Context) {
	var req createConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.create.Execute(c.Request.Context(), conversationapp.CreateConversationCommand{
		AgentID: req.AgentID, Title: req.Title,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{
		"conversationId": res.ConversationID,
		"agentId":        res.AgentID,
		"agentVersion":   res.AgentVersion,
	})
}

type startTurnRequest struct {
	Message string `json:"message" binding:"required"`
}

// StartTurn 开启一轮对话（仅记录用户消息，实际生成走 QA/分析用例）。
func (h *ConversationHandler) StartTurn(c *gin.Context) {
	var req startTurnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.start.Execute(c.Request.Context(), conversationapp.StartTurnCommand{
		ConversationID: c.Param("id"), Message: req.Message,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"conversationId": res.ConversationID, "turnId": res.TurnID})
}

// Compress 压缩会话上下文。
func (h *ConversationHandler) Compress(c *gin.Context) {
	res, err := h.compress.Execute(c.Request.Context(), conversationapp.CompressContextCommand{
		ConversationID: c.Param("id"),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{
		"conversationId": res.ConversationID,
		"summary":        res.Summary,
		"compressed":     res.Compressed,
	})
}
