// Package http 是接口层：把 HTTP 请求翻译成应用层用例调用。
//
// 本层只依赖应用层与领域层，不直接依赖基础设施层。
package http

import (
	"context"
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	analysisdomain "github.com/walry/ledger/internal/domain/analysis"
	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	reportdomain "github.com/walry/ledger/internal/domain/report"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
)

// FileStore 是接口层需要的最小文件存储能力（由基础设施层实现）。
type FileStore interface {
	Save(ctx context.Context, name string, data []byte) (string, error)
	Load(ctx context.Context, path string) ([]byte, error)
}

func respond(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func respondCreated(c *gin.Context, data any) { respond(c, nethttp.StatusCreated, data) }

// respondError 将错误映射为 HTTP 状态码。
func respondError(c *gin.Context, err error) {
	status := nethttp.StatusBadRequest
	if isNotFound(err) {
		status = nethttp.StatusNotFound
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func respondBindError(c *gin.Context, err error) {
	c.JSON(nethttp.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	for _, sentinel := range []error{
		reportdomain.ErrNotFound,
		knowledgedomain.ErrNotFound,
		agentdomain.ErrNotFound,
		conversationdomain.ErrNotFound,
		analysisdomain.ErrNotFound,
		tooldomain.ErrNotFound,
		modelconfigdomain.ErrNotFound,
	} {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}
