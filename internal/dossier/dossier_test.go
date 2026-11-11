package dossier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/classifier"
	"github.com/alexandrmotologa/guardmesh/internal/store"
)

func TestDossierGeneration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_dossier_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	meta := DossierMetadata{
		AppID:           "test-hiring-bot",
		SystemName:      "Enterprise Resume Screener",
		Organization:    "Acme Corp Europe",
		Author:          "Alexandr Motologa, Head of AI",
		IntendedPurpose: "Automated candidate pre-screening and CV qualification",
		ModelUsed:       "gpt-4o",
		GeneratedAt:     time.Now(),
		RiskEvaluation: classifier.RiskEvaluation{
			Tier:               classifier.TierHighRisk,
			Score:              0.85,
			Reasons:            []string{"Annex III 4(a): Recruitment and CV screening"},
			ApplicableArticles: []string{"Annex III(4)(a)", "Article 6", "Article 14"},
			RequiredControls:   []string{"Human Oversight (Art. 14)", "Technical Documentation (Art. 11)"},
			IsBlocked:          false,
		},
		TotalRequests:    1420,
		TotalPIIMasked:   538,
		ComplianceStatus: "NEEDS_REVIEW",
		SampleAuditRecords: []store.AuditEvent{
			{
				Timestamp:        time.Now().Add(-1 * time.Hour),
				Model:            "gpt-4o",
				LatencyMs:        4.2,
				RiskTier:         "HIGH_RISK",
				PIIDetectedCount: 2,
				IsBlocked:        false,
			},
		},
	}

	// Test Markdown
	md := GenerateAnnexIVMarkdown(meta)
	if !strings.Contains(md, "EU AI Act Technical Documentation Dossier") {
		t.Errorf("markdown missing title")
	}
	if !strings.Contains(md, "HIGH_RISK") {
		t.Errorf("markdown missing risk tier")
	}

	// Test NIST RMF
	nist := GenerateNISTRMFMarkdown(meta)
	if !strings.Contains(nist, "NIST AI Risk Management Framework") {
		t.Errorf("nist rmf markdown missing title")
	}

	// Test PDF
	pdfPath := filepath.Join(tempDir, "test_dossier.pdf")
	if err := GeneratePDF(meta, pdfPath); err != nil {
		t.Fatalf("generate pdf failed: %v", err)
	}

	fi, err := os.Stat(pdfPath)
	if err != nil {
		t.Fatalf("pdf file not created: %v", err)
	}
	if fi.Size() < 1000 {
		t.Errorf("pdf file too small, size: %d bytes", fi.Size())
	}
}
