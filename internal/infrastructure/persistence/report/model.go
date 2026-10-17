// Package report provides the GORM persistence adapter for the report bounded
// context. It implements the domain port report.Repository.
package report

import (
	"encoding/json"
	"time"

	domain "github.com/walry/ledger/internal/domain/report"
)

// ReportRow is the persistence model for the FinancialReport aggregate root.
type ReportRow struct {
	ID                string     `gorm:"column:id;primaryKey;type:text"`
	Name              string     `gorm:"column:name;type:text;not null"`
	SourceFilename    string     `gorm:"column:source_filename;type:text"`
	SourceContentType string     `gorm:"column:source_content_type;type:text"`
	SourceSize        int64      `gorm:"column:source_size;not null;default:0"`
	SourceStoragePath string     `gorm:"column:source_storage_path;type:text"`
	Status            string     `gorm:"column:status;type:text;not null;index"`
	RawText           string     `gorm:"column:raw_text;type:text"`
	FailureReason     string     `gorm:"column:failure_reason;type:text"`
	UploadedAt        time.Time  `gorm:"column:uploaded_at;not null"`
	ParsedAt          *time.Time `gorm:"column:parsed_at"`
	ArchivedAt        *time.Time `gorm:"column:archived_at"`
}

// TableName pins the physical table name.
func (ReportRow) TableName() string { return "report_reports" }

// SectionRow is the persistence model for the Section entity inside the
// report aggregate. It is stored as a child table with an ON DELETE CASCADE
// foreign key to report_reports.
type SectionRow struct {
	ID       string `gorm:"column:id;primaryKey;type:text"`
	ReportID string `gorm:"column:report_id;type:text;not null;index"`
	Type     string `gorm:"column:type;type:text;not null"`
	Title    string `gorm:"column:title;type:text"`
	Level    int    `gorm:"column:level;not null;default:0"`
	Content  string `gorm:"column:content;type:text"`
	Order    int    `gorm:"column:sort_order;not null;default:0"`
	Tables   []byte `gorm:"column:tables;type:jsonb"`

	// Report is only used by AutoMigrate to create the cascade FK; it is never
	// read or written by the repository (ReportID carries the relation).
	Report ReportRow `gorm:"foreignKey:ReportID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName pins the physical table name.
func (SectionRow) TableName() string { return "report_sections" }

// tableJSON is the JSONB wire form for report.Table.
type tableJSON struct {
	Caption string     `json:"caption"`
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

func tablesToJSON(tables []domain.Table) []byte {
	out := make([]tableJSON, 0, len(tables))
	for _, t := range tables {
		out = append(out, tableJSON{Caption: t.Caption(), Headers: t.Headers(), Rows: t.Rows()})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func tablesFromJSON(b []byte) []domain.Table {
	if len(b) == 0 {
		return make([]domain.Table, 0)
	}
	var in []tableJSON
	if err := json.Unmarshal(b, &in); err != nil {
		return make([]domain.Table, 0)
	}
	out := make([]domain.Table, 0, len(in))
	for _, t := range in {
		out = append(out, domain.NewTable(t.Caption, t.Headers, t.Rows))
	}
	return out
}

func reportToRow(r *domain.Report) *ReportRow {
	src := r.Source()
	return &ReportRow{
		ID:                r.ID().String(),
		Name:              r.Name(),
		SourceFilename:    src.Filename,
		SourceContentType: src.ContentType,
		SourceSize:        src.Size,
		SourceStoragePath: src.StoragePath,
		Status:            string(r.Status()),
		RawText:           r.RawText(),
		FailureReason:     r.FailureReason(),
		UploadedAt:        r.UploadedAt(),
		ParsedAt:          r.ParsedAt(),
		ArchivedAt:        r.ArchivedAt(),
	}
}

func sectionToRow(reportID string, s *domain.Section) *SectionRow {
	return &SectionRow{
		ID:       s.ID().String(),
		ReportID: reportID,
		Type:     string(s.Type()),
		Title:    s.Title(),
		Level:    s.Level(),
		Content:  s.Content(),
		Order:    s.Order(),
		Tables:   tablesToJSON(s.Tables()),
	}
}

func sectionFromRow(row *SectionRow) (*domain.Section, error) {
	s := domain.NewSection(domain.SectionId(row.ID), domain.SectionType(row.Type), row.Title, row.Level, row.Order)
	if row.Content != "" {
		s.SetContent(row.Content)
	}
	for _, t := range tablesFromJSON(row.Tables) {
		s.AddTable(t)
	}
	return s, nil
}

func reportFromRow(row *ReportRow, sections []*domain.Section) (*domain.Report, error) {
	src := domain.Source{
		Filename:    row.SourceFilename,
		ContentType: row.SourceContentType,
		Size:        row.SourceSize,
		StoragePath: row.SourceStoragePath,
	}
	return domain.Rehydrate(
		domain.ReportId(row.ID), row.Name, src, domain.Status(row.Status),
		sections, row.RawText, row.FailureReason,
		row.UploadedAt, row.ParsedAt, row.ArchivedAt,
	)
}
