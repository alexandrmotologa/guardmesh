package dossier

import (
	"fmt"
	"strings"
)

// GenerateNISTRMFMarkdown compiles NIST AI Risk Management Framework 1.0 cross-walk
func GenerateNISTRMFMarkdown(meta DossierMetadata) string {
	var sb strings.Builder

	sb.WriteString("# NIST AI Risk Management Framework (AI RMF 1.0) Cross-Walk\n\n")
	sb.WriteString(fmt.Sprintf("**System**: %s (`%s`) | **Date**: %s\n\n", meta.SystemName, meta.AppID, meta.GeneratedAt.Format("2006-01-02")))

	sb.WriteString("## 1. GOVERN (GV)\n")
	sb.WriteString("- **GV-1.1**: AI policies, processes, and legal compliance (EU AI Act & GDPR) are integrated into gateway routing.\n")
	sb.WriteString("- **GV-1.2**: Roles and responsibilities for AI risk management are assigned, with continuous audit trail logging.\n")
	sb.WriteString(fmt.Sprintf("- **GV-1.3**: Governance status currently evaluated as `%s` with %d recorded requests.\n\n", meta.ComplianceStatus, meta.TotalRequests))

	sb.WriteString("## 2. MAP (MP)\n")
	sb.WriteString(fmt.Sprintf("- **MP-1.1**: Intended purpose documented: \"%s\".\n", meta.IntendedPurpose))
	sb.WriteString(fmt.Sprintf("- **MP-2.1**: System risk tier mapped to `%s` based on automated prompt & tool analysis.\n", meta.RiskEvaluation.Tier))
	sb.WriteString("- **MP-3.1**: Third-party dependencies tracked (LLM endpoints, API versions, and request payload schema).\n\n")

	sb.WriteString("## 3. MEASURE (MS)\n")
	sb.WriteString(fmt.Sprintf("- **MS-1.1**: Sensitive data interception measured: %d PII entities masked to date.\n", meta.TotalPIIMasked))
	sb.WriteString("- **MS-2.1**: Adversarial robustness tested: Active prompt injection & jailbreak heuristic scanner enabled.\n")
	sb.WriteString("- **MS-3.1**: Latency impact and model execution times continuously tracked in SQLite audit store.\n\n")

	sb.WriteString("## 4. MANAGE (MG)\n")
	sb.WriteString("- **MG-1.1**: Zero-data retention on cloud providers enforced via cryptographic tokenization.\n")
	sb.WriteString("- **MG-2.1**: Emergency kill-switches and automated Article 5 hard blocks active.\n")
	sb.WriteString("- **MG-3.1**: Living technical documentation dossiers exportable on demand for internal and external auditors.\n")

	return sb.String()
}
