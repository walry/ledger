// Package modelconfig provides the GORM persistence adapter for the
// modelconfig bounded context. It implements the domain port
// modelconfig.Repository.
package modelconfig

import (
	"encoding/json"
	"time"

	domain "github.com/walry/ledger/internal/domain/modelconfig"
)

// ModelConfigRow is the persistence model for the ModelConfig aggregate.
type ModelConfigRow struct {
	ID        string    `gorm:"column:id;primaryKey;type:text"`
	Name      string    `gorm:"column:name;type:text;not null;uniqueIndex:ux_model_configs_name"`
	Provider  string    `gorm:"column:provider;type:text;not null"`
	BaseURL   string    `gorm:"column:base_url;type:text"`
	APIKey    string    `gorm:"column:api_key;type:text"`
	Model     string    `gorm:"column:model;type:text"`
	Params    []byte    `gorm:"column:parameters;type:jsonb"`
	Enabled   bool      `gorm:"column:enabled;not null;default:true"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName pins the physical table name.
func (ModelConfigRow) TableName() string { return "model_configs" }

func paramsToJSON(p domain.ModelParameters) []byte {
	b, err := json.Marshal(p)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func paramsFromJSON(b []byte) domain.ModelParameters {
	var p domain.ModelParameters
	if len(b) == 0 {
		return p
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return domain.ModelParameters{}
	}
	return p
}

func modelConfigToRow(m *domain.ModelConfig) *ModelConfigRow {
	return &ModelConfigRow{
		ID:        m.ID().String(),
		Name:      m.Name(),
		Provider:  string(m.Provider()),
		BaseURL:   m.BaseURL(),
		APIKey:    m.APIKey(),
		Model:     m.Model(),
		Params:    paramsToJSON(m.Parameters()),
		Enabled:   m.Enabled(),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}
}

func modelConfigFromRow(row *ModelConfigRow) *domain.ModelConfig {
	return domain.RehydrateModelConfig(
		domain.ModelConfigId(row.ID), row.Name, domain.ProviderType(row.Provider),
		row.BaseURL, row.APIKey, row.Model, paramsFromJSON(row.Params),
		row.Enabled, row.CreatedAt, row.UpdatedAt,
	)
}
