package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	analysisapp "github.com/walry/ledger/internal/application/analysis"
)

var errMissingFields = errors.New("missing required fields")

// AnalysisHandler 处理数字员工（分析任务）请求。
type AnalysisHandler struct {
	submit *analysisapp.SubmitTaskUseCase
	run    *analysisapp.RunTaskUseCase
	get    *analysisapp.GetResultUseCase
	list   *analysisapp.ListTasksUseCase
}

// NewAnalysisHandler 创建处理器。
func NewAnalysisHandler(
	submit *analysisapp.SubmitTaskUseCase,
	run *analysisapp.RunTaskUseCase,
	get *analysisapp.GetResultUseCase,
	list *analysisapp.ListTasksUseCase,
) *AnalysisHandler {
	return &AnalysisHandler{submit: submit, run: run, get: get, list: list}
}

// List 分页查询分析任务。
func (h *AnalysisHandler) List(c *gin.Context) {
	res, err := h.list.Execute(c.Request.Context(), analysisapp.ListTasksCommand{
		Offset: queryInt(c, "offset", 0),
		Limit:  queryInt(c, "limit", 20),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

type submitTaskRequest struct {
	Title     string   `json:"title"`
	Question  string   `json:"question" binding:"required"`
	AgentID   string   `json:"agentId" binding:"required"`
	ReportIDs []string `json:"reportIds"`
}

// Submit 提交分析任务（入队）。
func (h *AnalysisHandler) Submit(c *gin.Context) {
	var req submitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.submit.Execute(c.Request.Context(), analysisapp.SubmitTaskCommand{
		Title:     req.Title,
		Question:  req.Question,
		AgentID:   req.AgentID,
		ReportIDs: req.ReportIDs,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"taskId": res.TaskID})
}

// Run 同步执行分析任务。
func (h *AnalysisHandler) Run(c *gin.Context) {
	if err := h.run.Execute(c.Request.Context(), analysisapp.RunTaskCommand{TaskID: c.Param("id")}); err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"taskId": c.Param("id"), "status": "succeeded"})
}

// Get 查询分析结果。
func (h *AnalysisHandler) Get(c *gin.Context) {
	res, err := h.get.Execute(c.Request.Context(), analysisapp.GetResultCommand{TaskID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}
