package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/walry/ledger/internal/domain/modelconfig"
)

// streamBuffer is the size of the buffered chunk channel returned by Stream.
const streamBuffer = 16

// scanSSE reads a text/event-stream from r and invokes fn for every non-empty
// "data:" payload. The OpenAI "[DONE]" sentinel is skipped. It stops as soon as
// ctx is cancelled or fn returns an error.
func scanSSE(ctx context.Context, r io.Reader, fn func(data []byte) error) error {
	br := bufio.NewReaderSize(r, 64*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(trimmed, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if data != "" && data != "[DONE]" {
					if ferr := fn([]byte(data)); ferr != nil {
						return ferr
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// sendChunk delivers c on ch, aborting when ctx is cancelled.
func sendChunk(ctx context.Context, ch chan<- modelconfig.Chunk, c modelconfig.Chunk) bool {
	select {
	case ch <- c:
		return true
	case <-ctx.Done():
		return false
	}
}

// toolCallAccumulator assembles tool calls that streaming providers deliver in
// fragments (id/name on the first fragment, arguments split across many).
// Fragments are keyed by index so parallel calls interleave correctly.
type toolCallAccumulator struct {
	calls map[int]*modelconfig.ToolCall
	order []int
}

func newToolCallAccumulator() *toolCallAccumulator {
	return &toolCallAccumulator{calls: make(map[int]*modelconfig.ToolCall)}
}

// add appends a fragment for the tool call at index. Empty id/name/args are
// ignored so that later fragments do not clobber earlier values.
func (a *toolCallAccumulator) add(index int, id, name, args string) {
	c, ok := a.calls[index]
	if !ok {
		c = &modelconfig.ToolCall{}
		a.calls[index] = c
		a.order = append(a.order, index)
	}
	if id != "" {
		c.ID = id
	}
	if name != "" {
		c.Name = name
	}
	c.Arguments += args
}

// empty reports whether no fragments were accumulated.
func (a *toolCallAccumulator) empty() bool { return len(a.order) == 0 }

// result returns the assembled calls in first-seen order.
func (a *toolCallAccumulator) result() []modelconfig.ToolCall {
	out := make([]modelconfig.ToolCall, 0, len(a.order))
	for _, idx := range a.order {
		out = append(out, *a.calls[idx])
	}
	return out
}

// emitToolCalls sends each assembled call exactly once as a ToolCallDelta.
func emitToolCalls(ctx context.Context, ch chan<- modelconfig.Chunk, calls []modelconfig.ToolCall) bool {
	for i := range calls {
		call := calls[i]
		if !sendChunk(ctx, ch, modelconfig.Chunk{ToolCallDelta: &call}) {
			return false
		}
	}
	return true
}

// apiError is returned when a provider responds with a non-2xx status.
type apiError struct {
	Provider string
	Status   int
	Message  string
}

func (e *apiError) Error() string {
	if e.Status == 0 {
		return fmt.Sprintf("llm/%s: %s", e.Provider, e.Message)
	}
	return fmt.Sprintf("llm/%s: http %d: %s", e.Provider, e.Status, e.Message)
}

// decodeErrorMessage extracts a human-readable message from a provider error body.
func decodeErrorMessage(body []byte) string {
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Error) > 0 {
		var parsed struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(envelope.Error, &parsed); err == nil && parsed.Message != "" {
			return parsed.Message
		}
		var raw string
		if err := json.Unmarshal(envelope.Error, &raw); err == nil && raw != "" {
			return raw
		}
		return string(envelope.Error)
	}
	msg := strings.TrimSpace(string(body))
	if len(msg) > 512 {
		msg = msg[:512]
	}
	if msg == "" {
		msg = "unknown error"
	}
	return msg
}
