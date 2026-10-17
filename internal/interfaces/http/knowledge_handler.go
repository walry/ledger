package http

import (
	nethttp "net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	knowledgeapp "github.com/walry/ledger/internal/application/knowledge"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
)

// KnowledgeHandler 处理知识库相关请求。
type KnowledgeHandler struct {
	create  *knowledgeapp.CreateKnowledgeBaseUseCase
	addDoc  *knowledgeapp.AddDocumentUseCase
	reindex *knowledgeapp.ReindexDocumentUseCase
	search  *knowledgeapp.SearchKnowledgeUseCase
	listKB  *knowledgeapp.ListKnowledgeBasesUseCase
	getKB   *knowledgeapp.GetKnowledgeBaseUseCase
}

// NewKnowledgeHandler 创建处理器。
func NewKnowledgeHandler(
	create *knowledgeapp.CreateKnowledgeBaseUseCase,
	addDoc *knowledgeapp.AddDocumentUseCase,
	reindex *knowledgeapp.ReindexDocumentUseCase,
	search *knowledgeapp.SearchKnowledgeUseCase,
	listKB *knowledgeapp.ListKnowledgeBasesUseCase,
	getKB *knowledgeapp.GetKnowledgeBaseUseCase,
) *KnowledgeHandler {
	return &KnowledgeHandler{create: create, addDoc: addDoc, reindex: reindex, search: search, listKB: listKB, getKB: getKB}
}

// ListKnowledgeBases 分页查询知识库。
func (h *KnowledgeHandler) ListKnowledgeBases(c *gin.Context) {
	res, err := h.listKB.Execute(c.Request.Context(), knowledgeapp.ListKnowledgeBasesCommand{
		Offset: queryInt(c, "offset", 0),
		Limit:  queryInt(c, "limit", 50),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

// GetKnowledgeBase 查询知识库详情（含文档列表）。
func (h *KnowledgeHandler) GetKnowledgeBase(c *gin.Context) {
	res, err := h.getKB.Execute(c.Request.Context(), knowledgeapp.GetKnowledgeBaseCommand{KnowledgeBaseID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

type createKBRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	EmbeddingModel string `json:"embeddingModel"`
	Dimension      int    `json:"dimension"`
}

// CreateKnowledgeBase 创建知识库。
func (h *KnowledgeHandler) CreateKnowledgeBase(c *gin.Context) {
	var req createKBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.create.Execute(c.Request.Context(), knowledgeapp.CreateKnowledgeBaseCommand{
		Name:           req.Name,
		Description:    req.Description,
		EmbeddingModel: req.EmbeddingModel,
		Dimension:      req.Dimension,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"knowledgeBaseId": res.KnowledgeBaseID})
}

type addDocumentRequest struct {
	Title       string `json:"title"`
	SourceKind  string `json:"sourceKind"`
	SourceRefID string `json:"sourceRefId"`
	Text        string `json:"text" binding:"required"`
}

// AddDocument 向知识库新增文档并建立索引。
func (h *KnowledgeHandler) AddDocument(c *gin.Context) {
	var req addDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	kind := knowledgedomain.SourceKind(req.SourceKind)
	if kind == "" {
		kind = knowledgedomain.SourceFile
	}
	res, err := h.addDoc.Execute(c.Request.Context(), knowledgeapp.AddDocumentCommand{
		KnowledgeBaseID: c.Param("id"),
		Title:           req.Title,
		SourceKind:      kind,
		SourceRefID:     req.SourceRefID,
		Text:            req.Text,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"documentId": res.DocumentID, "chunkCount": res.ChunkCount})
}

type reindexRequest struct {
	Text string `json:"text"`
}

// Reindex 重建文档索引。
func (h *KnowledgeHandler) Reindex(c *gin.Context) {
	var req reindexRequest
	_ = c.ShouldBindJSON(&req)
	res, err := h.reindex.Execute(c.Request.Context(), knowledgeapp.ReindexDocumentCommand{
		DocumentID: c.Param("id"),
		Text:       req.Text,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"documentId": res.DocumentID, "chunkCount": res.ChunkCount})
}

type searchRequest struct {
	Query            string   `json:"query" binding:"required"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
	TopK             int      `json:"topK"`
}

// Search 检索知识库。
func (h *KnowledgeHandler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	res, err := h.search.Execute(c.Request.Context(), knowledgeapp.SearchQuery{
		Query:            req.Query,
		KnowledgeBaseIDs: req.KnowledgeBaseIDs,
		TopK:             req.TopK,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

func queryInt(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
