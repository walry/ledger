package toolhandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// 注册 pgx 驱动（driver name "pgx"）。pgx 已是工程内的既有依赖，
	// 因此此处不引入新的 module。
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/walry/ledger/internal/domain/tool"
)

const (
	sqlQueryTimeout = 10 * time.Second
	sqlMaxRows      = 100
)

var _ tool.Handler = (*SQLQueryHandler)(nil)

// SQLQueryHandler 是一个只读 SQL 处理器：未配置 DSN 时禁用。
//
// 仅在语句以 select 开头（去空白后、忽略大小写）时执行，并施加
// 短超时与行数上限；任何其他语句一律拒绝。
type SQLQueryHandler struct {
	db  *sql.DB
	err error
}

// NewSQLQueryHandler 创建只读 SQL 处理器。dsn 为空时返回一个始终提示
// “未配置”的处理器。
func NewSQLQueryHandler(dsn string) tool.Handler {
	h := &SQLQueryHandler{}
	if strings.TrimSpace(dsn) != "" {
		h.db, h.err = sql.Open("pgx", dsn)
	}
	return h
}

// Name 返回处理器名。
func (h *SQLQueryHandler) Name() string { return "sql_query" }

// Invoke 执行只读查询。
func (h *SQLQueryHandler) Invoke(ctx context.Context, args json.RawMessage) (tool.InvocationResult, error) {
	if err := ctx.Err(); err != nil {
		return tool.InvocationResult{}, err
	}
	if h.db == nil {
		return errorResult("sql query tool is not configured"), nil
	}
	if h.err != nil {
		return errorResult("sql query tool is unavailable: " + h.err.Error()), nil
	}

	var in struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return errorResult("invalid arguments: " + err.Error()), nil
	}
	stmt, err := sanitizeReadOnly(in.Query)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	qctx, cancel := context.WithTimeout(ctx, sqlQueryTimeout)
	defer cancel()

	rows, err := h.db.QueryContext(qctx, stmt)
	if err != nil {
		return errorResult("query failed: " + err.Error()), nil
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return errorResult("failed to read columns: " + err.Error()), nil
	}

	var b strings.Builder
	b.WriteString(strings.Join(cols, " | "))
	b.WriteString("\n")

	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	count := 0
	for rows.Next() {
		if count >= sqlMaxRows {
			b.WriteString("... [row limit reached]\n")
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			return errorResult("failed to scan row: " + err.Error()), nil
		}
		cells := make([]string, len(vals))
		for i, v := range vals {
			if v == nil {
				cells[i] = "NULL"
			} else if raw, ok := v.([]byte); ok {
				cells[i] = string(raw)
			} else {
				cells[i] = fmt.Sprintf("%v", v)
			}
		}
		b.WriteString(strings.Join(cells, " | "))
		b.WriteString("\n")
		count++
	}
	if err := rows.Err(); err != nil {
		return errorResult("query error: " + err.Error()), nil
	}
	return tool.InvocationResult{Content: b.String()}, nil
}

// sanitizeReadOnly 校验语句只读，并返回去除末尾分号后的语句。
func sanitizeReadOnly(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("query is required")
	}
	if strings.HasSuffix(s, ";") {
		s = strings.TrimSpace(strings.TrimSuffix(s, ";"))
	}
	if s == "" {
		return "", fmt.Errorf("query is required")
	}
	if strings.Contains(s, ";") {
		return "", fmt.Errorf("only a single statement is allowed")
	}
	trimmed := strings.ToLower(s)
	if !strings.HasPrefix(trimmed, "select") || !keywordBoundary(s, len("select")) {
		return "", fmt.Errorf("only read-only select statements are allowed")
	}
	return s, nil
}

// keywordBoundary 判断位置 i 处是否为关键字结束（避免把 "selective" 当作 select）。
func keywordBoundary(s string, i int) bool {
	if i >= len(s) {
		return true
	}
	c := s[i]
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
		return false
	}
	return true
}
