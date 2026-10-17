package http

import (
	"time"

	"github.com/gin-gonic/gin"

	qaapp "github.com/walry/ledger/internal/application/qa"
	runtimedomain "github.com/walry/ledger/internal/domain/runtime"
	"github.com/walry/ledger/internal/interfaces/http/sse"
)

// QAHandler 处理问答助手请求（含 SSE 流式）。
type QAHandler struct {
	ask       *qaapp.AskUseCase
	askStream *qaapp.AskStreamUseCase
}

// NewQAHandler 创建处理器。
func NewQAHandler(ask *qaapp.AskUseCase, askStream *qaapp.AskStreamUseCase) *QAHandler {
	return &QAHandler{ask: ask, askStream: askStream}
}

type askRequest struct {
	ConversationID   string   `json:"conversationId"`
	Question         string   `json:"question"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
}

// bindAsk 从 JSON body 或 query 读取参数。
func bindAsk(c *gin.Context) (askRequest, bool) {
	var req askRequest
	if c.Request.Method == "GET" {
		req.ConversationID = c.Query("conversationId")
		req.Question = c.Query("question")
		return req, req.ConversationID != "" && req.Question != ""
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return req, false
	}
	if req.ConversationID == "" || req.Question == "" {
		respondBindError(c, errMissingFields)
		return req, false
	}
	return req, true
}

// Ask 非流式问答。
func (h *QAHandler) Ask(c *gin.Context) {
	req, ok := bindAsk(c)
	if !ok {
		return
	}
	res, err := h.ask.Execute(c.Request.Context(), qaapp.AskCommand{
		ConversationID:   req.ConversationID,
		Question:         req.Question,
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, 200, res)
}

// AskStream 流式问答（SSE）。
func (h *QAHandler) AskStream(c *gin.Context) {
	req, ok := bindAsk(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	events, err := h.askStream.ExecuteStream(ctx, qaapp.AskCommand{
		ConversationID:   req.ConversationID,
		Question:         req.Question,
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	w := sse.NewWriter(c)
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if err := w.Comment("ping"); err != nil {
				return
			}
		case ev, open := <-events:
			if !open {
				_ = w.Send(sse.Event{Event: "done", Data: "[DONE]"})
				return
			}
			if err := streamEvent(w, ev); err != nil {
				return
			}
			if ev.Type == runtimedomain.EventDone || ev.Type == runtimedomain.EventError {
				return
			}
		}
	}
}

func streamEvent(w *sse.Writer, ev runtimedomain.StreamEvent) error {
	switch ev.Type {
	case runtimedomain.EventDelta:
		return w.Send(sse.Event{Event: "message", Data: gin.H{"delta": ev.Delta}})
	case runtimedomain.EventToolCall:
		return w.Send(sse.Event{Event: "tool_call", Data: gin.H{"id": ev.ToolCallID, "name": ev.ToolName}})
	case runtimedomain.EventToolResult:
		return w.Send(sse.Event{Event: "tool_result", Data: gin.H{
			"id": ev.ToolCallID, "name": ev.ToolName, "result": ev.ToolResult,
		}})
	case runtimedomain.EventError:
		msg := "unknown error"
		if ev.Err != nil {
			msg = ev.Err.Error()
		}
		return w.Send(sse.Event{Event: "error", Data: gin.H{"error": msg}})
	case runtimedomain.EventDone:
		return w.Send(sse.Event{Event: "done", Data: gin.H{
			"answer": ev.Answer,
			"usage":  gin.H{"promptTokens": ev.Usage.PromptTokens, "completionTokens": ev.Usage.CompletionTokens, "totalTokens": ev.Usage.TotalTokens},
		}})
	}
	return nil
}
