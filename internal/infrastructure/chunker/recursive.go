// Package chunker 提供 knowledge 领域 Chunker 端口的基础设施实现。
package chunker

import (
	"context"
	"math"
	"strconv"
	"strings"

	knowledgeservice "github.com/walry/ledger/internal/domain/knowledge/service"
)

const defaultMaxTokens = 512

var _ knowledgeservice.Chunker = (*RecursiveChunker)(nil)

// RecursiveChunker 按“段落 → 句子 → 字符”的顺序递归切分文本。
//
// Token 估算启发式（单一、固定）：
//   - 每个 CJK 字符（汉字/日文/韩文/全角标点）计 0.5 token；
//   - 其余字符计 0.25 token（约等于 4 字节/字符 → 1 token）。
//
// 最终对累计分值向上取整得到 Tokens。该启发式仅用于切分与预算控制，
// 不追求与具体分词器逐字对齐。
//
// 硬约束：Markdown 表格的连续行块（以 "|" 开头的连续行）永不被切开，
// 即使其本身超过 MaxTokens。
type RecursiveChunker struct{}

// NewRecursiveChunker 创建递归切片器。
func NewRecursiveChunker() *RecursiveChunker { return &RecursiveChunker{} }

// Chunk 将 text 切分为若干 RawChunk。
func (c *RecursiveChunker) Chunk(ctx context.Context, text string, opts knowledgeservice.ChunkOptions) ([]knowledgeservice.RawChunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if strings.TrimSpace(text) == "" {
		return []knowledgeservice.RawChunk{}, nil
	}

	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	overlap := opts.Overlap
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= maxTokens {
		overlap = maxTokens / 4
	}

	blocks := splitBlocks(text)

	chunks := make([]knowledgeservice.RawChunk, 0)
	var cur []string
	curTokens := 0

	emit := func() string {
		content := strings.TrimSpace(strings.Join(cur, "\n\n"))
		cur = nil
		curTokens = 0
		if content == "" {
			return ""
		}
		idx := len(chunks)
		chunks = append(chunks, knowledgeservice.RawChunk{
			Index:    idx,
			Content:  content,
			Tokens:   estimateTokens(content),
			Metadata: map[string]string{"index": strconv.Itoa(idx)},
		})
		return content
	}

	addBlock := func(block string) {
		block = strings.TrimSpace(block)
		if block == "" {
			return
		}
		bt := estimateTokens(block)
		if curTokens > 0 && curTokens+bt > maxTokens {
			prev := emit()
			if overlap > 0 && prev != "" {
				if tail := tailByTokens(prev, overlap); tail != "" {
					cur = []string{tail}
					curTokens = estimateTokens(tail)
				}
			}
		}
		cur = append(cur, block)
		curTokens += bt
	}

	for _, block := range blocks {
		if estimateTokens(block) > maxTokens && !isTableBlock(block) {
			for _, piece := range splitOversized(block, maxTokens) {
				addBlock(piece)
			}
			continue
		}
		addBlock(block)
	}
	if len(cur) > 0 {
		emit()
	}
	return chunks, nil
}

// splitBlocks 将文本切为块：连续 "|" 开头的行归为同一表格块；其余按空行分段。
func splitBlocks(text string) []string {
	lines := strings.Split(text, "\n")
	blocks := make([]string, 0)
	var cur []string
	inTable := false

	flush := func() {
		if len(cur) == 0 {
			return
		}
		blocks = append(blocks, strings.Join(cur, "\n"))
		cur = nil
	}

	for _, line := range lines {
		isTableRow := strings.HasPrefix(strings.TrimSpace(line), "|")
		switch {
		case isTableRow:
			if !inTable {
				flush()
				inTable = true
			}
			cur = append(cur, line)
		default:
			if inTable {
				flush()
				inTable = false
			}
			if strings.TrimSpace(line) == "" {
				flush()
				continue
			}
			cur = append(cur, line)
		}
	}
	flush()
	return blocks
}

func isTableBlock(block string) bool {
	for _, line := range strings.Split(block, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		return strings.HasPrefix(strings.TrimSpace(line), "|")
	}
	return false
}

// splitOversized 将过大的非表格块按句子切分，超长句子再按字符硬切。
func splitOversized(block string, maxTokens int) []string {
	var out []string
	var cur []string
	curTokens := 0
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.Join(cur, ""))
			cur = nil
			curTokens = 0
		}
	}
	for _, sentence := range splitSentences(block) {
		st := estimateTokens(sentence)
		if st > maxTokens {
			flush()
			out = append(out, hardSplit(sentence, maxTokens)...)
			continue
		}
		if curTokens > 0 && curTokens+st > maxTokens {
			flush()
		}
		cur = append(cur, sentence)
		curTokens += st
	}
	flush()
	return out
}

func splitSentences(s string) []string {
	var out []string
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(r)
		switch r {
		case '。', '！', '？', '!', '?', '；', ';', '.', '\n':
			out = append(out, b.String())
			b.Reset()
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

func hardSplit(s string, maxTokens int) []string {
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	var out []string
	var b strings.Builder
	score := 0.0
	for _, r := range s {
		b.WriteRune(r)
		score += runeTokenWeight(r)
		if score >= float64(maxTokens) {
			out = append(out, b.String())
			b.Reset()
			score = 0
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

// tailByTokens 返回 s 末尾约 overlapTokens 个 token 对应的子串。
func tailByTokens(s string, overlapTokens int) string {
	if overlapTokens <= 0 {
		return ""
	}
	runes := []rune(s)
	score := 0.0
	for i := len(runes) - 1; i >= 0; i-- {
		score += runeTokenWeight(runes[i])
		if score >= float64(overlapTokens) {
			return string(runes[i:])
		}
	}
	return s
}

func estimateTokens(s string) int {
	return int(math.Ceil(tokenScore(s)))
}

func tokenScore(s string) float64 {
	score := 0.0
	for _, r := range s {
		score += runeTokenWeight(r)
	}
	return score
}

func runeTokenWeight(r rune) float64 {
	if isCJK(r) {
		return 0.5
	}
	return 0.25
}

func isCJK(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF: // CJK 统一表意
		return true
	case r >= 0x3400 && r <= 0x4DBF: // 扩展 A
		return true
	case r >= 0x3000 && r <= 0x303F: // CJK 标点
		return true
	case r >= 0x3040 && r <= 0x30FF: // 日文假名
		return true
	case r >= 0xAC00 && r <= 0xD7AF: // 韩文
		return true
	case r >= 0xFF00 && r <= 0xFFEF: // 全角
		return true
	default:
		return false
	}
}
