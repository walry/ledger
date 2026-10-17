package report

import "strings"

// Table 是报告中的表格（值对象）。
type Table struct {
	caption string
	headers []string
	rows    [][]string
}

// NewTable 创建表格。headers 与 rows 会被防御性拷贝。
func NewTable(caption string, headers []string, rows [][]string) Table {
	h := make([]string, len(headers))
	copy(h, headers)
	r := make([][]string, len(rows))
	for i, row := range rows {
		cp := make([]string, len(row))
		copy(cp, row)
		r[i] = cp
	}
	return Table{caption: caption, headers: h, rows: r}
}

// Caption 返回表标题。
func (t Table) Caption() string { return t.caption }

// Headers 返回表头。
func (t Table) Headers() []string {
	out := make([]string, len(t.headers))
	copy(out, t.headers)
	return out
}

// Rows 返回数据行。
func (t Table) Rows() [][]string {
	out := make([][]string, len(t.rows))
	for i, row := range t.rows {
		cp := make([]string, len(row))
		copy(cp, row)
		out[i] = cp
	}
	return out
}

// Empty 判断表格是否为空。
func (t Table) Empty() bool { return len(t.headers) == 0 && len(t.rows) == 0 }

// Markdown 将表格渲染为 Markdown，供检索上下文与提示词使用。
func (t Table) Markdown() string {
	var b strings.Builder
	if t.caption != "" {
		b.WriteString(t.caption)
		b.WriteString("\n")
	}
	if len(t.headers) > 0 {
		b.WriteString("| ")
		b.WriteString(strings.Join(t.headers, " | "))
		b.WriteString(" |\n| ")
		sep := make([]string, len(t.headers))
		for i := range sep {
			sep[i] = "---"
		}
		b.WriteString(strings.Join(sep, " | "))
		b.WriteString(" |\n")
	}
	for _, row := range t.rows {
		b.WriteString("| ")
		b.WriteString(strings.Join(row, " | "))
		b.WriteString(" |\n")
	}
	return b.String()
}
