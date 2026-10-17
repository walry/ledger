// Package idgen 定义标识符生成端口。
//
// 端口定义在领域侧（被领域工厂/聚合消费），实现在基础设施层
// （见 internal/infrastructure/idgen），从而让领域层不依赖任何第三方 SDK。
package idgen

// Generator 是唯一标识符生成端口。
type Generator interface {
	// New 返回一个新的、全局唯一的字符串标识符。
	New() string
}
