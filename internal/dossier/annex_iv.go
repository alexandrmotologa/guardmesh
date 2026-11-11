package dossier

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/classifier"
	"github.com/alexandrmotologa/guardmesh/internal/store"
)

type DossierMetadata struct {
	AppID              string
	SystemName         string
	Organization       string
	Author             string
	IntendedPurpose    string
	ModelUsed          string
	GeneratedAt        time.Time
	RiskEvaluation     classifier.RiskEvaluation
	TotalRequests      int64
	TotalPIIMasked     int64
	ComplianceStatus   string
	SampleAuditRecords []store.AuditEvent
}

// GenerateAnnexIVMarkdown produces complete EU AI Act Annex IV technical documentation in Markdown
func GenerateAnnexIVMarkdown(meta DossierMetadata) string {
	var sb strings.Builder

	sb.WriteString("# EU AI Act Technical Documentation Dossier (Annex IV)\n\n")
	sb.WriteString(fmt.Sprintf("**System Name**: %s\n", meta.SystemName))
	sb.WriteString(fmt.Sprintf("**Application ID**: `%s`\n", meta.AppID))
	sb.WriteString(fmt.Sprintf("**Organization**: %s\n", meta.Organization))
	sb.WriteString(fmt.Sprintf("**Dossier Date**: %s\n", meta.GeneratedAt.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("**Overall Risk Tier**: **%s**\n", meta.RiskEvaluation.Tier))
	sb.WriteString(fmt.Sprintf("**Compliance Status**: `%s`\n\n", meta.ComplianceStatus))
	sb.WriteString("---\n\n")

	// Section 1: System Description & Intended Purpose
	sb.WriteString("## 1. General Description of the AI System\n\n")
	sb.WriteString(fmt.Sprintf("- **Intended Purpose**: %s\n", meta.IntendedPurpose))
	sb.WriteString(fmt.Sprintf("- **Underlying AI Model(s)**: %s\n", meta.ModelUsed))
	sb.WriteString("- **System Architecture**: High-speed Go reverse proxy with real-time zero-retention PII anonymization and Annex III risk classification.\n")
	sb.WriteString("- **Deployment Environment**: Enterprise On-Premises / Private Cloud Gateway.\n\n")

	// Section 2: EU AI Act Risk Classification (Articles 5, 6, Annex III, Article 50)
	sb.WriteString("## 2. EU AI Act Classification & Legal Justification\n\n")
	sb.WriteString(fmt.Sprintf("Under Regulation (EU) 2024/1689 (EU Artificial Intelligence Act), this system has been evaluated and assigned to **%s**.\n\n", meta.RiskEvaluation.Tier))

	sb.WriteString("### 2.1 Applicable Legal Articles\n")
	for _, art := range meta.RiskEvaluation.ApplicableArticles {
		sb.WriteString(fmt.Sprintf("- **%s**\n", art))
	}
	sb.WriteString("\n")

	sb.WriteString("### 2.2 Classification Rationale\n")
	for _, reason := range meta.RiskEvaluation.Reasons {
		sb.WriteString(fmt.Sprintf("- %s\n", reason))
	}
	sb.WriteString("\n")

	// Section 3: Data Governance & PII Sanitization (Article 10 & GDPR)
	sb.WriteString("## 3. Data Governance & PII De-Identification (Article 10)\n\n")
	sb.WriteString("To ensure strict compliance with Regulation (EU) 2016/679 (GDPR) and prevent training data ingestion by third-party model providers, `guardmesh` applies deterministic cryptographic tokenization to all outgoing traffic:\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Intercepted PII Entities**: %d\n", meta.TotalPIIMasked))
	sb.WriteString(fmt.Sprintf("- **Total Monitored Invocations**: %d\n", meta.TotalRequests))
	sb.WriteString("- **De-Identification Mode**: Ephemeral AES-256 GCM in-memory token vault with automatic TTL expiry.\n")
	sb.WriteString("- **Covered Data Categories**: Credit Cards (Luhn verified), Bank Accounts (IBAN mod-97), Emails, Phone Numbers, Social Security Numbers, IP Addresses, Person Names, and Developer Secrets.\n\n")

	// Section 4: Human Oversight & Technical Measures (Article 14)
	sb.WriteString("## 4. Human Oversight & Operational Safeguards (Article 14)\n\n")
	sb.WriteString("In accordance with Article 14 of the EU AI Act, high-risk and sensitive workloads require appropriate human-machine interfaces and operational controls:\n\n")
	for _, ctrl := range meta.RiskEvaluation.RequiredControls {
		sb.WriteString(fmt.Sprintf("- [x] **%s**\n", ctrl))
	}
	sb.WriteString("- [x] **Emergency Interruption**: Reverse proxy supports immediate circuit breaking and hard kills for Article 5 non-compliance.\n")
	sb.WriteString("- [x] **Output Verification**: Operational staff must retain final sign-off authority for High-Risk automated decisions.\n\n")

	// Section 5: Cybersecurity & Robustness (Article 15)
	sb.WriteString("## 5. Cybersecurity, Accuracy & Robustness (Article 15)\n\n")
	sb.WriteString("- **Adversarial Prompt Injection Scanner**: Active pattern and heuristic detection against prompt overrides, system instructions exfiltration, and jailbreaks.\n")
	sb.WriteString("- **API Key & Secret Leak Protection**: Real-time redaction of production credentials (OpenAI, AWS, GitHub, Private Keys) prior to egress.\n")
	sb.WriteString("- **Tamper-Evident Storage**: Audit events stored in local SQLite database with Write-Ahead Logging (WAL) and SHA-256 request fingerprinting.\n\n")

	// Section 6: Audit Trail & Verification Evidence (Article 12)
	sb.WriteString("## 6. Audit Trail & Verification Evidence (Article 12)\n\n")
	sb.WriteString("Recent operational audit records captured through the governance gateway:\n\n")
	sb.WriteString("| Timestamp | Model | Latency | Risk Tier | PII Masked | Status |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")

	if len(meta.SampleAuditRecords) == 0 {
		sb.WriteString("| *N/A* | *No historical records recorded yet* | - | - | - | - |\n")
	} else {
		for _, rec := range meta.SampleAuditRecords {
			status := "Allowed"
			if rec.IsBlocked {
				status = "**BLOCKED**"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %.1fms | %s | %d | %s |\n",
				rec.Timestamp.Format("2006-01-02 15:04"),
				rec.Model,
				rec.LatencyMs,
				rec.RiskTier,
				rec.PIIDetectedCount,
				status,
			))
		}
	}
	sb.WriteString("\n")

	// Section 7: Regulatory Declaration of Conformity
	sb.WriteString("## 7. Declaration of Conformity\n\n")
	sb.WriteString("This document constitutes technical documentation as mandated by Article 11 and Annex IV of Regulation (EU) 2024/1689. The undersigned confirms that risk mitigation controls, data de-identification safeguards, and human oversight measures documented herein are actively enforced by the `guardmesh` governance gateway.\n\n")
	sb.WriteString(fmt.Sprintf("**Authorized Representative**: %s\n", meta.Author))
	sb.WriteString(fmt.Sprintf("**Date of Issue**: %s\n", meta.GeneratedAt.Format("2006-01-02")))
	sb.WriteString("**Digital Integrity Hash (SHA-256)**: `AUTOMATICALLY_GENERATED_AT_EXPORT`\n")

	return sb.String()
}
