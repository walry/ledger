package knowledge

import (
	"errors"
	"math"
)

// Embedding 是文本向量（值对象）。
type Embedding struct {
	vector []float32
	model  string
}

// NewEmbedding 创建向量并校验不变量。
func NewEmbedding(model string, vector []float32) (Embedding, error) {
	if len(vector) == 0 {
		return Embedding{}, errors.New("knowledge: embedding vector is empty")
	}
	cp := make([]float32, len(vector))
	copy(cp, vector)
	return Embedding{vector: cp, model: model}, nil
}

// Vector 返回向量副本。
func (e Embedding) Vector() []float32 {
	out := make([]float32, len(e.vector))
	copy(out, e.vector)
	return out
}

// Model 返回产生该向量的模型名。
func (e Embedding) Model() string { return e.model }

// Dimension 返回向量维度。
func (e Embedding) Dimension() int { return len(e.vector) }

// IsZero 判断向量是否为空。
func (e Embedding) IsZero() bool { return len(e.vector) == 0 }

// Cosine 计算与另一个向量的余弦相似度。
func (e Embedding) Cosine(other Embedding) float64 {
	if len(e.vector) == 0 || len(e.vector) != len(other.vector) {
		return 0
	}
	var dot, na, nb float64
	for i := range e.vector {
		a := float64(e.vector[i])
		b := float64(other.vector[i])
		dot += a * b
		na += a * a
		nb += b * b
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
