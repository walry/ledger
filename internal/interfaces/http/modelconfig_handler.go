package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	modelconfigapp "github.com/walry/ledger/internal/application/modelconfig"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
)

// ModelConfigHandler 处理模型配置相关请求。
type ModelConfigHandler struct {
	create *modelconfigapp.CreateModelConfigUseCase
	update *modelconfigapp.UpdateModelConfigUseCase
	list   *modelconfigapp.ListModelConfigsUseCase
	get    *modelconfigapp.GetModelConfigUseCase
}

// NewModelConfigHandler 创建处理器。
func NewModelConfigHandler(
	create *modelconfigapp.CreateModelConfigUseCase,
	update *modelconfigapp.UpdateModelConfigUseCase,
	list *modelconfigapp.ListModelConfigsUseCase,
	get *modelconfigapp.GetModelConfigUseCase,
) *ModelConfigHandler {
	return &ModelConfigHandler{create: create, update: update, list: list, get: get}
}

// List 分页查询模型配置。
func (h *ModelConfigHandler) List(c *gin.Context) {
	res, err := h.list.Execute(c.Request.Context(), modelconfigapp.ListModelConfigsCommand{
		Offset: queryInt(c, "offset", 0),
		Limit:  queryInt(c, "limit", 50),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

// Get 查询模型配置详情。
func (h *ModelConfigHandler) Get(c *gin.Context) {
	res, err := h.get.Execute(c.Request.Context(), modelconfigapp.GetModelConfigCommand{ConfigID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

type createModelConfigRequest struct {
	Name       string                            `json:"name" binding:"required"`
	Provider   string                            `json:"provider" binding:"required"`
	BaseURL    string                            `json:"baseUrl"`
	APIKey     string                            `json:"apiKey"`
	Model      string                            `json:"model" binding:"required"`
	Parameters modelconfigdomain.ModelParameters `json:"parameters"`
}

// Create 创建模型配置。
func (h *ModelConfigHandler) Create(c *gin.Context) {
	var req createModelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	params := req.Parameters
	if params.MaxTokens == 0 && params.Temperature == 0 && params.TopP == 0 {
		params = modelconfigdomain.DefaultParameters()
	}
	res, err := h.create.Execute(c.Request.Context(), modelconfigapp.CreateModelConfigCommand{
		Name:       req.Name,
		Provider:   modelconfigdomain.ProviderType(req.Provider),
		BaseURL:    req.BaseURL,
		APIKey:     req.APIKey,
		Model:      req.Model,
		Parameters: params,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"configId": res.ConfigID})
}

type updateModelConfigRequest struct {
	BaseURL    string                            `json:"baseUrl"`
	APIKey     string                            `json:"apiKey"`
	Model      string                            `json:"model" binding:"required"`
	Parameters modelconfigdomain.ModelParameters `json:"parameters"`
}

// Update 更新模型配置。
func (h *ModelConfigHandler) Update(c *gin.Context) {
	var req updateModelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	params := req.Parameters
	if params.MaxTokens == 0 && params.Temperature == 0 && params.TopP == 0 {
		params = modelconfigdomain.DefaultParameters()
	}
	res, err := h.update.Execute(c.Request.Context(), modelconfigapp.UpdateModelConfigCommand{
		ConfigID:   c.Param("id"),
		BaseURL:    req.BaseURL,
		APIKey:     req.APIKey,
		Model:      req.Model,
		Parameters: params,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"configId": res.ConfigID})
}
