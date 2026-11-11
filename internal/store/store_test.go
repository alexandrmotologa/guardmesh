package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_AuditAndApps(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_store_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	// Record audit event
	event := &AuditEvent{
		ID:               "evt-1",
		Timestamp:        time.Now(),
		AppID:            "crm-portal",
		ClientIP:         "192.168.1.10",
		Provider:         "openai",
		Model:            "gpt-4o",
		StatusCode:       200,
		LatencyMs:        3.4,
		PromptTokens:     120,
		CompletionTokens: 80,
		RiskTier:         "HIGH_RISK",
		RiskScore:        0.75,
		RiskReasons:      []string{"Annex III 4(a): Recruitment or CV selection"},
		IsBlocked:        false,
		PIIDetectedCount: 3,
		PIICategories:    []string{"EMAIL", "PERSON_NAME"},
		PromptSHA256:     "abcdef123456",
		SampleSnippet:    "Screening candidate Alice Doe...",
	}

	if err := s.RecordAuditEvent(event); err != nil {
		t.Fatalf("record audit event failed: %v", err)
	}

	// Query events
	events, err := s.QueryAuditEvents(AuditFilter{AppID: "crm-portal"})
	if err != nil {
		t.Fatalf("query audit events failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "evt-1" {
		t.Errorf("expected ID evt-1, got %s", events[0].ID)
	}
	if len(events[0].RiskReasons) != 1 {
		t.Errorf("expected 1 risk reason, got %d", len(events[0].RiskReasons))
	}

	// Check application registry
	apps, err := s.ListApplications()
	if err != nil {
		t.Fatalf("list applications failed: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].ID != "crm-portal" || apps[0].HighestRiskTier != "HIGH_RISK" {
		t.Errorf("unexpected app state: %+v", apps[0])
	}

	// Check global stats
	stats, err := s.GetGlobalStats()
	if err != nil {
		t.Fatalf("get global stats failed: %v", err)
	}
	if stats.TotalRequests != 1 || stats.TotalPIIMasked != 3 {
		t.Errorf("unexpected stats: %+v", stats)
	}
}
