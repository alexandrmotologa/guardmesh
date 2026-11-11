# guardmesh Windows PowerShell Quickstart Demo Script
# Tests health checks, classification, PII scanning, audit logs, and dossier downloads.

$ErrorActionPreference = "Stop"
$GuardmeshUrl = if ($env:GUARDMESH_URL) { $env:GUARDMESH_URL } else { "http://localhost:8080" }

Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  guardmesh Windows PowerShell API Verification" -ForegroundColor Cyan
Write-Host "  Gateway: $GuardmeshUrl" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan

# 1. Health & Readiness
Write-Host "`n[1] Checking Health & Readiness..." -ForegroundColor Yellow
$health = Invoke-RestMethod -Uri "$GuardmeshUrl/healthz" -Method Get
Write-Host "Health Check: $($health.status)" -ForegroundColor Green
$ready = Invoke-RestMethod -Uri "$GuardmeshUrl/readyz" -Method Get
Write-Host "Readiness Check: $($ready.status)" -ForegroundColor Green

# 2. PII Scanner
Write-Host "`n[2] Testing PII & Secret Scanner Endpoint (/api/scan)..." -ForegroundColor Yellow
$scanBody = @{
    text = "Transfer $50,000 to Alice Walker with Romanian CNP 1800514123459, Spanish NIF 12345678Z, IBAN DE89370400440532013000, email alice@domain.com, and badge EMP-823901."
} | ConvertTo-Json

$scanResult = Invoke-RestMethod -Uri "$GuardmeshUrl/api/scan" -Method Post -Body $scanBody -ContentType "application/json"
Write-Host "Sanitized Text: $($scanResult.sanitized)" -ForegroundColor DarkGray
Write-Host "Entities Detected: $($scanResult.entities.Count)" -ForegroundColor Green
foreach ($ent in $scanResult.entities) {
    Write-Host "  - [$($ent.type)] '$($ent.value)' -> '$($ent.token)' (confidence: $($ent.confidence))" -ForegroundColor Cyan
}

# 3. EU AI Act Classifier
Write-Host "`n[3] Testing EU AI Act Risk Classifier Endpoint (/api/classify)..." -ForegroundColor Yellow
$highRiskBody = @{
    app_id = "hr-recruiting-engine"
    prompt = "Evaluate applicant resumes, score candidate suitability, and automatically filter out lower percentile job applicants."
} | ConvertTo-Json

$classifyHigh = Invoke-RestMethod -Uri "$GuardmeshUrl/api/classify" -Method Post -Body $highRiskBody -ContentType "application/json"
Write-Host "High-Risk Prompt Tier: $($classifyHigh.risk_tier) (Category: $($classifyHigh.category))" -ForegroundColor Magenta
Write-Host "Article 5 Prohibited: $($classifyHigh.article5_prohibited), Annex III: $($classifyHigh.annex_iii_high_risk)"

$prohibitedBody = @{
    app_id = "security-monitor"
    prompt = "Use biometric emotion recognition to monitor employee engagement and stress levels during meetings."
} | ConvertTo-Json

$classifyProhibited = Invoke-RestMethod -Uri "$GuardmeshUrl/api/classify" -Method Post -Body $prohibitedBody -ContentType "application/json"
Write-Host "Prohibited Prompt Tier: $($classifyProhibited.risk_tier) (Article 5 Prohibited: $($classifyProhibited.article5_prohibited))" -ForegroundColor Red

# 4. Inventory & Stats
Write-Host "`n[4] Querying Application Inventory & Stats..." -ForegroundColor Yellow
$stats = Invoke-RestMethod -Uri "$GuardmeshUrl/api/stats" -Method Get
Write-Host "Total Events: $($stats.total_events), Total Blocked: $($stats.total_blocked), PII Intercepted: $($stats.total_pii_intercepted)" -ForegroundColor Green

# 5. Exporting Telemetry
Write-Host "`n[5] Testing Audit Log Exporter..." -ForegroundColor Yellow
$csv = Invoke-RestMethod -Uri "$GuardmeshUrl/api/events/export?format=csv" -Method Get
$csvLines = $csv -split "`n"
Write-Host "Exported CSV records count: $($csvLines.Count)" -ForegroundColor Green

# 6. Dossier Export
Write-Host "`n[6] Downloading Technical Compliance Dossiers..." -ForegroundColor Yellow
$mdOut = "$env:TEMP\dossier_default.md"
$pdfOut = "$env:TEMP\dossier_default.pdf"

Invoke-WebRequest -Uri "$GuardmeshUrl/api/apps/default/dossier/md" -OutFile $mdOut
Write-Host "Markdown dossier saved to $mdOut" -ForegroundColor Green

Invoke-WebRequest -Uri "$GuardmeshUrl/api/apps/default/dossier/pdf" -OutFile $pdfOut
Write-Host "PDF dossier saved to $pdfOut" -ForegroundColor Green

Write-Host "`n======================================================================" -ForegroundColor Cyan
Write-Host "  All API endpoints verified successfully." -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
