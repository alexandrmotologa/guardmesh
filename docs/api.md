# guardmesh API Reference

The `guardmesh` gateway exposes standard LLM proxy interfaces along with REST endpoints for governance telemetry, audit log exports, observability, and compliance dossiers.

---

## 1. Proxy Interfaces

### `POST /v1/chat/completions`
OpenAI compatible endpoint. Accepts standard OpenAI request payloads.

#### Custom Request Headers
- `X-Guardmesh-App-ID`: String identifier for the caller application (default: `default-app`). Used to group metrics and compile application-specific dossiers.
- `Authorization`: Bearer token forwarded directly to upstream OpenAI.

#### Custom Response Headers
- `X-Guardmesh-Risk-Tier`: Evaluated risk tier (`PROHIBITED`, `HIGH_RISK`, `LIMITED_RISK`, `MINIMAL_RISK`).
- `X-Guardmesh-PII-Masked`: Integer count of sensitive entities de-identified in the request.
- `X-Guardmesh-Blocked`: Present and set to `"true"` if the request was halted by compliance rules.
- `X-AI-Generated`: Present and set to `"true"` when Article 50 Transparency is active.
- `X-EU-AI-Act-Article`: `"Article 50(1)"` when transparency disclosure applies.

#### Example Request
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-proj-upstream-key" \
  -H "X-Guardmesh-App-ID: crm-bot" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "Process refund for John Doe, Romanian CNP 1800514123459, IBAN DE89370400440532013000."}
    ],
    "stream": false
  }'
```

---

### `POST /v1/messages`
Anthropic Claude compatible endpoint. Accepts standard Anthropic request payloads (`messages`, `system`, `max_tokens`, `stream`).

#### Example Request
```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: anthropic-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "X-Guardmesh-App-ID: support-agent" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 1024,
    "system": "You are a customer support agent.",
    "messages": [
      {"role": "user", "content": "Customer contact is alice@example.com, NIF 12345678Z."}
    ]
  }'
```

---

## 2. Governance & Compliance Endpoints

### `GET /api/stats`
Returns global proxy metrics and risk tier distributions.

#### Example Response
```json
{
  "total_requests": 1420,
  "total_blocked": 12,
  "total_pii_masked": 492,
  "avg_latency_ms": 3.42,
  "risk_distribution": {
    "MINIMAL_RISK": 1100,
    "LIMITED_RISK": 220,
    "HIGH_RISK": 88,
    "PROHIBITED": 12
  }
}
```

---

### `GET /api/events`
Query recent audit log records.

#### Query Parameters
- `limit`: Number of records to return (default: 50).
- `offset`: Pagination offset.
- `app_id`: Filter by application identifier.
- `risk_tier`: Filter by tier (`PROHIBITED`, `HIGH_RISK`, `LIMITED_RISK`, `MINIMAL_RISK`).
- `is_blocked`: Boolean flag (`true` or `false`).

---

### `GET /api/events/export`
Exports audit event logs for external SIEM or regulatory compliance review.

#### Query Parameters
- `format`: Export format (`csv` or `jsonl`, default: `csv`).
- `app_id`: Optional filter by application identifier.
- `risk_tier`: Optional filter by risk tier.
- `is_blocked`: Optional filter by block state.

#### Example Request
```bash
curl "http://localhost:8080/api/events/export?format=csv&risk_tier=HIGH_RISK" -o high_risk_audit.csv
```

---

### `GET /api/apps`
Returns an array of registered internal applications discovered through gateway traffic.

---

### `POST /api/classify`
Direct audit endpoint to evaluate any system prompt or tool context for EU AI Act compliance.

#### Request Body
```json
{
  "app_id": "candidate-evaluator",
  "text": "Screen resumes and rank job applicants for junior developer role."
}
```

---

### `POST /api/scan`
Direct PII detection endpoint. Returns identified entities and sanitized output.

#### Request Body
```json
{
  "text": "Employee EMP-482019 with Romanian CNP 1800514123459, Spanish NIF 12345678Z, email alice@domain.com."
}
```

---

### `GET /api/apps/{app_id}/dossier/{format}`
Generates and downloads an EU AI Act Annex IV technical documentation dossier on demand.

- Format `pdf`: Returns binary application/pdf.
- Format `md`: Returns text/markdown file.

---

## 3. Observability & Health Endpoints

### `GET /healthz`
Health probe for container runtimes and Kubernetes liveness probes.
```json
{"status":"healthy","version":"1.0.0","time":"2026-11-11T16:00:00Z"}
```

### `GET /readyz`
Readiness probe asserting database connectivity and vault operational readiness.
```json
{"status":"ready"}
```

### `GET /metrics`
Standard Prometheus metrics endpoint.

#### Exported Metrics
- `guardmesh_requests_total`: Total number of requests processed.
- `guardmesh_blocked_requests_total`: Total prohibited or blocked requests.
- `guardmesh_pii_entities_masked_total`: Total personal data entities intercepted.
- `guardmesh_average_latency_ms`: Average proxy overhead latency.
- `guardmesh_risk_distribution_total{tier="..."}`: Count by EU AI Act risk tier.
