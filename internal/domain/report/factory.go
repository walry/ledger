package report

import (
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// Factory 负责创建 Report 聚合（封装 ID 生成与时间来源）。
type Factory struct {
	ids   idgen.Generator
	clock clock.Clock
}

// NewFactory 创建工厂。
func NewFactory(ids idgen.Generator, clk clock.Clock) *Factory {
	return &Factory{ids: ids, clock: clk}
}

// Create 新建一份报告。
func (f *Factory) Create(name string, source Source) (*Report, error) {
	return NewReport(ReportId(f.ids.New()), name, source, f.clock.Now())
}
