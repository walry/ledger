// Package clock 抽象时间来源，便于在测试中控制"当前时间"。
package clock

import "time"

// Clock 抽象时间来源。
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// System 返回基于系统时钟的实现。
func System() Clock { return systemClock{} }

// Fixed 返回一个始终返回固定时间的实现，便于测试。
func Fixed(t time.Time) Clock { return fixedClock{t: t} }

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }
