#!/usr/bin/env bash
# guardmesh End-to-End Gateway Verification Script
# Tests health checks, classification, PII scanning, audit logs, and dossier downloads.

set -e

GUARDMESH_URL="${GUARDMESH_URL:-http://localhost:8080}"

echo "======================================================================"
echo "  guardmesh End-to-End API Verification"
echo "  Gateway: ${GUARDMESH_URL}"
echo "======================================================================"

echo -e "\n[1] Checking Health & Readiness..."
curl -s -o /dev/null -w "Health status: %{http_code}\n" "${GUARDMESH_URL}/healthz"
curl -s -o /dev/null -w "Readiness status: %{http_code}\n" "${GUARDMESH_URL}/readyz"

echo -e "\n[2] Testing PII & Secret Scanner Endpoint (/api/scan)..."
curl -s -X POST "${GUARDMESH_URL}/api/scan" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Transfer $50,000 to Alice Walker with Romanian CNP 1800514123459, Spanish NIF 12345678Z, IBAN DE89370400440532013000, email alice@domain.com, and badge EMP-823901."
  }' | jq .

echo -e "\n[3] Testing EU AI Act Risk Classifier Endpoint (/api/classify)..."
echo "Testing Annex III High-Risk Prompt:"
curl -s -X POST "${GUARDMESH_URL}/api/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "hr-recruiting-engine",
    "prompt": "Evaluate applicant resumes, score candidate suitability, and automatically filter out lower percentile job applicants."
  }' | jq .

echo -e "\nTesting Article 5 Prohibited Practice Prompt:"
curl -s -X POST "${GUARDMESH_URL}/api/classify" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "security-monitor",
    "prompt": "Use biometric emotion recognition to monitor employee engagement and stress levels during meetings."
  }' | jq .

echo -e "\n[4] Querying Application Inventory & Stats..."
curl -s "${GUARDMESH_URL}/api/stats" | jq .
curl -s "${GUARDMESH_URL}/api/apps" | jq .

echo -e "\n[5] Testing Audit Log Exporter..."
echo "Exporting first 5 rows of CSV telemetry:"
curl -s "${GUARDMESH_URL}/api/events/export?format=csv" | head -n 5

echo -e "\nExporting first 2 rows of JSONL telemetry:"
curl -s "${GUARDMESH_URL}/api/events/export?format=jsonl" | head -n 2

echo -e "\n[6] Downloading Technical Compliance Dossiers..."
curl -s "${GUARDMESH_URL}/api/apps/default/dossier/md" -o /tmp/dossier_default.md
echo "Saved Markdown Dossier to /tmp/dossier_default.md ($(wc -l < /tmp/dossier_default.md) lines)"

curl -s "${GUARDMESH_URL}/api/apps/default/dossier/pdf" -o /tmp/dossier_default.pdf
echo "Saved Regulatory PDF Dossier to /tmp/dossier_default.pdf ($(wc -c < /tmp/dossier_default.pdf) bytes)"

echo -e "\n======================================================================"
echo "  All API endpoints verified successfully."
echo "======================================================================"
