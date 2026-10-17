package toolhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/walry/ledger/internal/domain/tool"
)

var _ tool.Handler = (*CalculatorHandler)(nil)

// CalculatorHandler 安全求值基础算术表达式，不依赖 eval、也不 shell 调用。
type CalculatorHandler struct{}

// NewCalculatorHandler 创建计算器处理器。
func NewCalculatorHandler() *CalculatorHandler { return &CalculatorHandler{} }

// Name 返回处理器名。
func (h *CalculatorHandler) Name() string { return "calculator" }

// Invoke 求值 args.expression，支持 + - * / % ( ) 与数字。
func (h *CalculatorHandler) Invoke(ctx context.Context, args json.RawMessage) (tool.InvocationResult, error) {
	if err := ctx.Err(); err != nil {
		return tool.InvocationResult{}, err
	}
	var in struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return errorResult("invalid arguments: " + err.Error()), nil
	}
	if strings.TrimSpace(in.Expression) == "" {
		return errorResult("expression is required"), nil
	}
	v, err := evalExpression(in.Expression)
	if err != nil {
		return errorResult(err.Error()), nil
	}
	return tool.InvocationResult{Content: formatNumber(v)}, nil
}

func errorResult(msg string) tool.InvocationResult {
	return tool.InvocationResult{Content: msg, IsError: true}
}

func formatNumber(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// evalExpression 使用递归下降解析求值表达式。
func evalExpression(s string) (float64, error) {
	p := &exprParser{s: s}
	v, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if p.pos != len(p.s) {
		return 0, fmt.Errorf("unexpected character %q at position %d", p.s[p.pos], p.pos)
	}
	return v, nil
}

type exprParser struct {
	s   string
	pos int
}

func (p *exprParser) skipSpace() {
	for p.pos < len(p.s) && (p.s[p.pos] == ' ' || p.s[p.pos] == '\t') {
		p.pos++
	}
}

func (p *exprParser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			break
		}
		op := p.s[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *exprParser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			break
		}
		op := p.s[p.pos]
		if op != '*' && op != '/' && op != '%' {
			break
		}
		p.pos++
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		switch op {
		case '*':
			left *= right
		case '/':
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		case '%':
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		}
	}
	return left, nil
}

func (p *exprParser) parseFactor() (float64, error) {
	p.skipSpace()
	if p.pos >= len(p.s) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	switch c := p.s[p.pos]; {
	case c == '+':
		p.pos++
		return p.parseFactor()
	case c == '-':
		p.pos++
		v, err := p.parseFactor()
		return -v, err
	case c == '(':
		p.pos++
		v, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		p.skipSpace()
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return v, nil
	default:
		return p.parseNumber()
	}
}

func (p *exprParser) parseNumber() (float64, error) {
	p.skipSpace()
	start := p.pos
	seenDot := false
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		switch {
		case c >= '0' && c <= '9':
			p.pos++
		case c == '.' && !seenDot:
			seenDot = true
			p.pos++
		default:
			goto done
		}
	}
done:
	if start == p.pos {
		return 0, fmt.Errorf("expected number at position %d", start)
	}
	v, err := strconv.ParseFloat(p.s[start:p.pos], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", p.s[start:p.pos])
	}
	return v, nil
}
