package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	toolapp "github.com/walry/ledger/internal/application/tool"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
)

// ToolHandler 处理工具注册相关请求。
type ToolHandler struct {
	register *toolapp.RegisterToolUseCase
	enable   *toolapp.EnableToolUseCase
	list     *toolapp.ListToolsUseCase
	get      *toolapp.GetToolUseCase
}

// NewToolHandler 创建处理器。
func NewToolHandler(
	register *toolapp.RegisterToolUseCase,
	enable *toolapp.EnableToolUseCase,
	list *toolapp.ListToolsUseCase,
	get *toolapp.GetToolUseCase,
) *ToolHandler {
	return &ToolHandler{register: register, enable: enable, list: list, get: get}
}

// List 分页查询工具。
func (h *ToolHandler) List(c *gin.Context) {
	res, err := h.list.Execute(c.Request.Context(), toolapp.ListToolsCommand{
		Offset: queryInt(c, "offset", 0),
		Limit:  queryInt(c, "limit", 20),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

// Get 查询工具详情。
func (h *ToolHandler) Get(c *gin.Context) {
	res, err := h.get.Execute(c.Request.Context(), toolapp.GetToolCommand{ToolID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

type registerToolRequest struct {
	Name           string                `json:"name" binding:"required"`
	Description    string                `json:"description"`
	HandlerName    string                `json:"handlerName" binding:"required"`
	Schema         tooldomain.JSONSchema `json:"schema" binding:"required"`
	SourceKind     string                `json:"sourceKind"`
	SourceURI      string                `json:"sourceUri"`
	SourceConfig   map[string]string     `json:"sourceConfig"`
	TimeoutSeconds int                   `json:"timeoutSeconds"`
}

// Register 注册工具。
func (h *ToolHandler) Register(c *gin.Context) {
	var req registerToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	kind := tooldomain.SourceKind(req.SourceKind)
	if kind == "" {
		kind = tooldomain.SourceBuiltin
	}
	res, err := h.register.Execute(c.Request.Context(), toolapp.RegisterToolCommand{
		Name:           req.Name,
		Description:    req.Description,
		HandlerName:    req.HandlerName,
		Schema:         req.Schema,
		SourceKind:     kind,
		SourceURI:      req.SourceURI,
		SourceConfig:   req.SourceConfig,
		TimeoutSeconds: req.TimeoutSeconds,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"toolId": res.ToolID})
}

type enableToolRequest struct {
	Enabled bool `json:"enabled"`
}

// Enable 启用/停用工具。
func (h *ToolHandler) Enable(c *gin.Context) {
	var req enableToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.enable.Execute(c.Request.Context(), toolapp.EnableToolCommand{
		ToolID: c.Param("id"), Enabled: req.Enabled,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"toolId": res.ToolID, "enabled": res.Enabled})
}
