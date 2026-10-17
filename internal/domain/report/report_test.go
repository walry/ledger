package report

import (
	"testing"
	"time"
)

type testIDs struct{ n int }

func (t *testIDs) New() string { t.n++; return "s-" + string(rune('0'+t.n)) }

func TestReportLifecycle(t *testing.T) {
	now := time.Date(2026, 2, 1, 9, 0, 0, 0, time.UTC)
	ids := &testIDs{}

	r, err := NewReport(NewReportId("r-1"), "2025年度报告", Source{
		Filename: "2025.pdf", ContentType: "application/pdf", Size: 1024, StoragePath: "/tmp/2025.pdf",
	}, now)
	if err != nil {
		t.Fatalf("new report: %v", err)
	}
	if r.Status() != StatusUploaded {
		t.Fatalf("expected uploaded, got %s", r.Status())
	}
	events := r.PullEvents()
	if len(events) != 1 || events[0].EventName() != "ReportUploaded" {
		t.Fatalf("expected ReportUploaded event, got %v", events)
	}

	if err := r.MarkParsing(); err != nil {
		t.Fatalf("mark parsing: %v", err)
	}
	doc := ParsedDocument{
		RawText: "第一章 营业收入",
		Sections: []ParsedSection{
			{Type: SectionChapter, Title: "第一章", Level: 1, Content: "营业收入"},
			{Type: SectionTable, Title: "利润表", Tables: []Table{
				NewTable("利润表", []string{"项目", "金额"}, [][]string{{"营业收入", "42"}}),
			}},
		},
	}
	if err := r.MarkParsed(doc, ids, now); err != nil {
		t.Fatalf("mark parsed: %v", err)
	}
	if r.Status() != StatusParsed {
		t.Fatalf("expected parsed, got %s", r.Status())
	}
	if len(r.Sections()) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(r.Sections()))
	}
	if got := r.Sections()[1].Tables(); len(got) != 1 || got[0].Caption() != "利润表" {
		t.Fatalf("expected profit table, got %v", got)
	}
	if events := r.PullEvents(); len(events) != 1 || events[0].EventName() != "ReportParsed" {
		t.Fatalf("expected ReportParsed event, got %v", events)
	}

	if err := r.Archive(now); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if err := r.MarkParsed(doc, ids, now); err == nil {
		t.Fatalf("expected error re-parsing an archived report")
	}
}

func TestReportMarkFailedThenReparse(t *testing.T) {
	now := time.Now()
	r, _ := NewReport(NewReportId("r-2"), "报告", Source{}, now)
	r.MarkFailed("解析器不支持", now)
	if r.Status() != StatusFailed || r.FailureReason() == "" {
		t.Fatalf("expected failed status with reason")
	}
	if err := r.MarkParsing(); err != nil {
		t.Fatalf("failed report should be re-parsable: %v", err)
	}
}
