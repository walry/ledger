// Package tool provides the GORM persistence adapter for the tool bounded
// context. It implements the domain port tool.Repository.
package tool

import (
	"encoding/json"
	"time"

	domain "github.com/walry/ledger/internal/domain/tool"
)

// ToolRow is the persistence model for the Tool aggregate. The JSON schema and
// the source descriptor are stored as JSONB columns.
type ToolRow struct {
	ID          string    `gorm:"column:id;primaryKey;type:text"`
	Name        string    `gorm:"column:name;type:text;not null;uniqueIndex:ux_tools_name"`
	Description string    `gorm:"column:description;type:text"`
	Schema      []byte    `gorm:"column:schema;type:jsonb"`
	SourceKind  string    `gorm:"column:source_kind;type:text"`
	SourceURI   string    `gorm:"column:source_uri;type:text"`
	SourceConf  []byte    `gorm:"column:source_config;type:jsonb"`
	HandlerName string    `gorm:"column:handler_name;type:text"`
	Enabled     bool      `gorm:"column:enabled;not null;default:true"`
	Version     int       `gorm:"column:version;not null;default:1"`
	TimeoutNS   int64     `gorm:"column:timeout_ns;not null;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
}

// TableName pins the physical table name.
func (ToolRow) TableName() string { return "tools" }

func marshalJSON(v any, fallback string) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(fallback)
	}
	return b
}

func schemaFromJSON(b []byte) domain.JSONSchema {
	var s domain.JSONSchema
	if len(b) == 0 {
		return s
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return domain.JSONSchema{}
	}
	return s
}

func sourceConfigFromJSON(b []byte) map[string]string {
	out := map[string]string{}
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]string{}
	}
	if out == nil {
		out = map[string]string{}
	}
	return out
}

func toolToRow(t *domain.Tool) *ToolRow {
	src := t.Source()
	config := src.Config()
	if config == nil {
		config = map[string]string{}
	}
	return &ToolRow{
		ID:          t.ID().String(),
		Name:        t.Name(),
		Description: t.Description(),
		Schema:      marshalJSON(t.Schema(), `{"type":"object"}`),
		SourceKind:  string(src.Kind()),
		SourceURI:   src.URI(),
		SourceConf:  marshalJSON(config, "{}"),
		HandlerName: t.Handler().Name(),
		Enabled:     t.Enabled(),
		Version:     t.Version(),
		TimeoutNS:   int64(t.Timeout()),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
}

func toolFromRow(row *ToolRow) *domain.Tool {
	source, err := domain.NewToolSource(domain.SourceKind(row.SourceKind), row.SourceURI, sourceConfigFromJSON(row.SourceConf))
	if err != nil {
		source = domain.ToolSource{}
	}
	handler, err := domain.NewHandlerRef(row.HandlerName)
	if err != nil {
		handler = domain.HandlerRef{}
	}
	return domain.RehydrateTool(
		domain.ToolId(row.ID), row.Name, row.Description, schemaFromJSON(row.Schema), source,
		handler, row.Enabled, row.Version, time.Duration(row.TimeoutNS), row.CreatedAt, row.UpdatedAt,
	)
}
