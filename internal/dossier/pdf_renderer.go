package dossier

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/go-pdf/fpdf"
)

// GeneratePDF creates an official EU AI Act Annex IV PDF report
func GeneratePDF(meta DossierMetadata, outputPath string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Document Title & Metadata
	pdf.SetTitle(fmt.Sprintf("EU AI Act Dossier - %s", meta.AppID), true)
	pdf.SetAuthor(meta.Author, true)

	// Header & Footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pageText := fmt.Sprintf("Page %d | guardmesh EU AI Act Compliance Engine | Confidential Dossier", pdf.PageNo())
		pdf.CellFormat(0, 10, pageText, "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	// --- Header Banner ---
	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(30, 41, 59) // Slate-800
	pdf.Cell(0, 10, "EU AI ACT TECHNICAL DOSSIER")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(100, 116, 139) // Slate-500
	pdf.Cell(0, 6, "Mandatory Technical Documentation pursuant to Article 11 and Annex IV of Regulation (EU) 2024/1689")
	pdf.Ln(10)

	// Horizontal Rule
	pdf.SetDrawColor(203, 213, 225)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// --- Executive Metadata Grid ---
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(71, 85, 105)
	pdf.Cell(40, 6, "Application ID:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, meta.AppID)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "Generated Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, meta.GeneratedAt.Format("2006-01-02 15:04"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "System Name:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(50, 6, meta.SystemName)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "Compliance Status:")
	pdf.SetFont("Arial", "B", 10)
	if meta.ComplianceStatus == "COMPLIANT" {
		pdf.SetTextColor(22, 163, 74) // Green
	} else if meta.ComplianceStatus == "NEEDS_REVIEW" {
		pdf.SetTextColor(217, 119, 6) // Amber
	} else {
		pdf.SetTextColor(220, 38, 38) // Red
	}
	pdf.Cell(40, 6, meta.ComplianceStatus)
	pdf.Ln(10)

	// --- Risk Classification Badge ---
	pdf.SetFont("Arial", "B", 12)
	switch meta.RiskEvaluation.Tier {
	case "PROHIBITED":
		pdf.SetFillColor(254, 242, 242)
		pdf.SetDrawColor(239, 68, 68)
		pdf.SetTextColor(185, 28, 28)
	case "HIGH_RISK":
		pdf.SetFillColor(254, 243, 199)
		pdf.SetDrawColor(245, 158, 11)
		pdf.SetTextColor(180, 83, 9)
	case "LIMITED_RISK":
		pdf.SetFillColor(239, 246, 255)
		pdf.SetDrawColor(59, 130, 246)
		pdf.SetTextColor(29, 78, 216)
	default:
		pdf.SetFillColor(240, 253, 244)
		pdf.SetDrawColor(34, 197, 94)
		pdf.SetTextColor(21, 128, 61)
	}

	badgeText := fmt.Sprintf("  RISK CLASSIFICATION: %s  (Score: %.2f)", meta.RiskEvaluation.Tier, meta.RiskEvaluation.Score)
	pdf.Rect(20, pdf.GetY(), 170, 10, "DF")
	pdf.CellFormat(170, 10, badgeText, "", 0, "L", false, 0, "")
	pdf.Ln(14)

	// --- Section 1: System Purpose & Architecture ---
	pdf.SetTextColor(30, 41, 59)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "1. General Description & Intended Purpose")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(51, 65, 85)
	purposeText := fmt.Sprintf("Intended Purpose: %s\nUnderlying Model(s): %s\nDeployment Architecture: On-premises reverse proxy with real-time zero-retention PII anonymization and Annex III risk classification.", meta.IntendedPurpose, meta.ModelUsed)
	pdf.MultiCell(170, 5, purposeText, "", "L", false)
	pdf.Ln(4)

	// --- Section 2: Legal Classification & Reasons ---
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 41, 59)
	pdf.Cell(0, 8, "2. Legal Justification & Applicable Articles")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(51, 65, 85)
	if len(meta.RiskEvaluation.ApplicableArticles) > 0 {
		pdf.Cell(0, 5, fmt.Sprintf("Applicable Articles: %s", meta.RiskEvaluation.ApplicableArticles))
		pdf.Ln(6)
	}

	for _, reason := range meta.RiskEvaluation.Reasons {
		pdf.Cell(5, 5, "- ")
		pdf.MultiCell(165, 5, reason, "", "L", false)
		pdf.Ln(1)
	}
	pdf.Ln(4)

	// --- Section 3: Data Governance & PII Metrics ---
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 41, 59)
	pdf.Cell(0, 8, "3. Data Governance & PII De-Identification (Article 10)")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(51, 65, 85)
	govText := fmt.Sprintf("Total Request Invocations Monitored: %d\nTotal Personal Data (PII) Entities Intercepted: %d\nCryptographic Token Vault: Ephemeral AES-256 GCM in-memory mappings with automated TTL expiry.\nZero Data Retention Guarantee: External cloud models receive solely de-identified replacement tokens.", meta.TotalRequests, meta.TotalPIIMasked)
	pdf.MultiCell(170, 5, govText, "", "L", false)
	pdf.Ln(6)

	// --- Page 2: Operational Controls & Audit Trail ---
	pdf.AddPage()

	// Section 4: Human Oversight
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 41, 59)
	pdf.Cell(0, 8, "4. Human Oversight & Operational Safeguards (Article 14)")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(51, 65, 85)
	for _, ctrl := range meta.RiskEvaluation.RequiredControls {
		pdf.Cell(6, 5, "[X]")
		pdf.MultiCell(164, 5, fmt.Sprintf(" %s", ctrl), "", "L", false)
		pdf.Ln(1)
	}
	pdf.Cell(6, 5, "[X]")
	pdf.MultiCell(164, 5, " Emergency Interruption: Real-time circuit breaker & hard kill switch for Article 5 non-compliance.", "", "L", false)
	pdf.Ln(6)

	// Section 5: Cybersecurity
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 41, 59)
	pdf.Cell(0, 8, "5. Cybersecurity & System Robustness (Article 15)")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(51, 65, 85)
	secText := "- Active Adversarial Scanner: Blocks prompt injection, jailbreaks, and instructions exfiltration.\n- Secret & Credential Redaction: Strips API keys (OpenAI, AWS, GitHub, Private Keys) before egress.\n- Immutable SQLite Audit Store: High-throughput WAL logging with SHA-256 fingerprinting."
	pdf.MultiCell(170, 5, secText, "", "L", false)
	pdf.Ln(6)

	// Section 6: Audit Trail Sample Table
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 41, 59)
	pdf.Cell(0, 8, "6. Audit Trail & Verification Telemetry (Article 12)")
	pdf.Ln(7)

	// Table Header
	pdf.SetFillColor(241, 245, 249) // Slate-100
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(71, 85, 105)
	pdf.CellFormat(35, 6, "Timestamp", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 6, "Model", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 6, "Latency", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 6, "Risk Tier", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 6, "PII Masked", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 6, "Status", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(51, 65, 85)
	if len(meta.SampleAuditRecords) == 0 {
		pdf.CellFormat(170, 6, "No historical requests recorded in database", "1", 1, "C", false, 0, "")
	} else {
		for i, rec := range meta.SampleAuditRecords {
			if i >= 6 {
				break
			}
			status := "Allowed"
			if rec.IsBlocked {
				status = "BLOCKED"
			}
			pdf.CellFormat(35, 5, rec.Timestamp.Format("2006-01-02 15:04"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 5, rec.Model, "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 5, fmt.Sprintf("%.1f ms", rec.LatencyMs), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 5, rec.RiskTier, "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 5, fmt.Sprintf("%d", rec.PIIDetectedCount), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 5, status, "1", 1, "C", false, 0, "")
		}
	}
	pdf.Ln(8)

	// Section 7: Declaration of Conformity
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(30, 41, 59)
	pdf.Cell(0, 8, "7. Regulatory Declaration of Conformity")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(51, 65, 85)
	declText := "The undersigned confirms that the technical documentation herein represents an authentic and complete record of governance, risk classification, and data privacy safeguards enforced by guardmesh."
	pdf.MultiCell(170, 4, declText, "", "L", false)
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(50, 5, fmt.Sprintf("Authorized Lead: %s", meta.Author))
	pdf.Cell(50, 5, fmt.Sprintf("Organization: %s", meta.Organization))
	pdf.Ln(8)

	// Hash stamp calculation
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s-%d", meta.AppID, meta.RiskEvaluation.Tier, meta.GeneratedAt.String(), meta.TotalRequests)))
	hashStr := hex.EncodeToString(h.Sum(nil))

	pdf.SetFont("Courier", "", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.Cell(0, 5, fmt.Sprintf("Digital Audit Fingerprint (SHA-256): %s", hashStr))

	return pdf.OutputFileAndClose(outputPath)
}

// GeneratePDFBytes generates the PDF and returns its raw bytes
func GeneratePDFBytes(meta DossierMetadata) ([]byte, error) {
	tempFile, err := os.CreateTemp("", "guardmesh_dossier_*.pdf")
	if err != nil {
		return nil, err
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	if err := GeneratePDF(meta, tempPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tempPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WriteDossierPackage writes both PDF and Markdown files to the destination
func WriteDossierPackage(meta DossierMetadata, outDir string) (pdfPath, mdPath string, err error) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", "", fmt.Errorf("create dossier out dir: %w", err)
	}

	mdPath = fmt.Sprintf("%s/EU_AI_Act_Annex_IV_%s.md", outDir, meta.AppID)
	pdfPath = fmt.Sprintf("%s/EU_AI_Act_Annex_IV_%s.pdf", outDir, meta.AppID)

	mdContent := GenerateAnnexIVMarkdown(meta)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return "", "", fmt.Errorf("write markdown dossier: %w", err)
	}

	if err := GeneratePDF(meta, pdfPath); err != nil {
		return "", "", fmt.Errorf("generate pdf dossier: %w", err)
	}

	return pdfPath, mdPath, nil
}
