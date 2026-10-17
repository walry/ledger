// Package idgen 提供 pkg/idgen.Generator 端口的基础设施实现。
package idgen

import (
	"github.com/google/uuid"

	pkidgen "github.com/walry/ledger/pkg/idgen"
)

var _ pkidgen.Generator = UUIDGenerator{}

// UUIDGenerator 基于 github.com/google/uuid 生成 v4 UUID。
type UUIDGenerator struct{}

// NewUUIDGenerator 创建 UUID 生成器。
func NewUUIDGenerator() UUIDGenerator { return UUIDGenerator{} }

// New 返回一个新的 v4 UUID 字符串。
func (UUIDGenerator) New() string { return uuid.NewString() }
