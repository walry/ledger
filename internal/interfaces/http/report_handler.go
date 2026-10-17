package http

import (
	"io"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	reportapp "github.com/walry/ledger/internal/application/report"
)

// ReportHandler 处理财务报告相关请求。
type ReportHandler struct {
	upload *reportapp.UploadReportUseCase
	parse  *reportapp.ParseReportUseCase
	get    *reportapp.GetReportUseCase
	list   *reportapp.ListReportsUseCase
	files  FileStore
}

// NewReportHandler 创建处理器。
func NewReportHandler(
	upload *reportapp.UploadReportUseCase,
	parse *reportapp.ParseReportUseCase,
	get *reportapp.GetReportUseCase,
	list *reportapp.ListReportsUseCase,
	files FileStore,
) *ReportHandler {
	return &ReportHandler{upload: upload, parse: parse, get: get, list: list, files: files}
}

// Upload 上传报告（multipart/form-data，字段名 file，可选 name）。
func (h *ReportHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, err)
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		respondError(c, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		respondError(c, err)
		return
	}
	name := c.PostForm("name")
	if name == "" {
		name = fileHeader.Filename
	}
	ctx := c.Request.Context()
	path, err := h.files.Save(ctx, fileHeader.Filename, data)
	if err != nil {
		respondError(c, err)
		return
	}
	res, err := h.upload.Execute(ctx, reportapp.UploadReportCommand{
		Name:        name,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Data:        data,
		StoragePath: path,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{"reportId": res.ReportID})
}

// Parse 解析已上传的报告（从存储读取原始文件）。
func (h *ReportHandler) Parse(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	detail, err := h.get.Execute(ctx, reportapp.GetReportCommand{ReportID: id})
	if err != nil {
		respondError(c, err)
		return
	}
	var data []byte
	if detail.StoragePath != "" {
		data, err = h.files.Load(ctx, detail.StoragePath)
		if err != nil {
			respondError(c, err)
			return
		}
	}
	res, err := h.parse.Execute(ctx, reportapp.ParseReportCommand{
		ReportID:    id,
		Data:        data,
		Filename:    detail.Filename,
		ContentType: detail.ContentType,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, gin.H{"reportId": res.ReportID, "sectionCount": res.SectionCount})
}

// Get 查询报告详情。
func (h *ReportHandler) Get(c *gin.Context) {
	res, err := h.get.Execute(c.Request.Context(), reportapp.GetReportCommand{ReportID: c.Param("id")})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}

// List 分页查询报告。
func (h *ReportHandler) List(c *gin.Context) {
	res, err := h.list.Execute(c.Request.Context(), reportapp.ListReportsCommand{
		Status: c.Query("status"),
		Offset: queryInt(c, "offset", 0),
		Limit:  queryInt(c, "limit", 20),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respond(c, nethttp.StatusOK, res)
}
