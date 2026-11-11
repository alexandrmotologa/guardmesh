<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="guardmesh Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">guardmesh</h1>

<p align="center">
  <strong>Enterprise AI Governance Gateway, Real-Time Bidirectional PII Anonymizer & Automated EU AI Act Compliance Auditor</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react" alt="React 19" />
  <img src="https://img.shields.io/badge/Regulation-EU%202024%2F1689-003399?style=flat" alt="EU AI Act" />
  <img src="https://img.shields.io/badge/GDPR-Compliant-success?style=flat" alt="GDPR" />
  <img src="https://img.shields.io/badge/Latency-<0.1ms-emerald?style=flat" alt="Latency" />
  <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License" />
</p>

<p align="center">
  <img src="docs/images/guardmesh_demo.gif?raw=true" alt="guardmesh Live Execution Demo" width="800" style="border-radius: 12px; box-shadow: 0 8px 30px rgba(0,0,0,0.3);" />
</p>

---

Deploying generative AI within corporate infrastructure introduces regulatory liabilities under **Regulation (EU) 2024/1689 (EU AI Act)** and **GDPR**. Violations of prohibited practices carry statutory fines up to €35,000,000 or 7% of global annual turnover. Internal microservices routinely transmit customer personal data, developer secrets, and internal documents to third-party model providers.

`guardmesh` acts as an on-premises reverse proxy between internal applications and external model providers (OpenAI, Anthropic). It strips personal data and secrets before outbound transmission, reconstructs original values in streaming responses, classifies workloads into EU AI Act risk tiers, blocks prohibited practices, enforces Article 50 transparency notices, and generates regulator-ready technical documentation dossiers in PDF and Markdown.

---

## Key Capabilities

- **Zero-Code SDK Integration**: Compatible with OpenAI and Anthropic API endpoints. Redirect calls by updating the base URL (for example, `http://localhost:8080/v1`).
- **Bidirectional PII De-Identification**: Scans for credit cards (Luhn verified), bank accounts (IBAN mod-97 verified), Romanian CNP (key 279146358279 algorithm), Spanish NIF/NIE (modulo 23 check), emails, phone numbers, SSNs, IP addresses, person names, and developer API keys.
- **Custom Enterprise DLP Rules**: Define custom regular expression patterns in configuration (for example, proprietary employee badges like `EMP-[0-9]{6}`).
- **Streaming Token Boundary Buffer**: Resolves tokens split across Server-Sent Events (SSE) packet boundaries without increasing stream latency.
- **Article 5 Hard-Kill Enforcer**: Rejects prohibited AI practices (such as workplace emotion recognition, social scoring, or biometric categorization) with an HTTP 403 Forbidden status code.
- **Article 50 Transparency Enforcer**: Automatically injects mandatory transparency notices and `X-AI-Generated: true` headers for customer-facing conversational assistants.
- **Real-Time Webhook Alerting**: Dispatches structured alert notifications to Slack, Microsoft Teams, or SIEM endpoints when prohibited practices or prompt injections occur.
- **Annex III High-Risk Classification**: Evaluates system instructions and tool definitions against Annex III domains (recruitment, credit assessment, essential public services, education).
- **1-Click Audit Dossier Generation**: Exports complete Annex IV technical documentation and NIST AI RMF cross-walks in PDF and Markdown format.
- **Audit Log Exporter**: Exports historical audit telemetry in CSV and JSONL formats for external regulatory reviews.
- **Cloud & Kubernetes Observability**: Native `/healthz`, `/readyz`, and Prometheus `/metrics` endpoints.
- **Embedded Compliance Cockpit**: Serves a real-time React 19 dashboard directly from the single Go binary at `http://localhost:8080/`.
- **Sub-Millisecond Overhead**: Operates with an average proxy overhead of under 0.1 milliseconds per request.

---

## Visual Interface & Architecture

### Desktop Executive Cockpit
<p align="center">
  <img src="docs/images/guardmesh_overview.png?raw=true" alt="guardmesh Executive Dashboard" width="800" style="border-radius: 10px;" />
</p>

### Live Audit Telemetry & Forensic Event Inspector
<p align="center">
  <img src="docs/images/guardmesh_events.png?raw=true" alt="guardmesh Live Audit Log" width="800" style="border-radius: 10px;" />
</p>

<p align="center">
  <img src="docs/images/guardmesh_event_modal.png?raw=true" alt="guardmesh Audit Event Forensic Modal" width="800" style="border-radius: 10px;" />
</p>

### EU AI Act Pre-Deployment Auditor & 4-Stage PII Simulator
<p align="center">
  <img src="docs/images/guardmesh_auditor.png?raw=true" alt="guardmesh EU AI Act Auditor" width="800" style="border-radius: 10px;" />
</p>

<p align="center">
  <img src="docs/images/guardmesh_simulator.png?raw=true" alt="guardmesh PII Simulator" width="800" style="border-radius: 10px;" />
</p>

### Discovered Applications & Shadow AI Registry (1-Click Annex IV PDF Export)
<p align="center">
  <img src="docs/images/guardmesh_apps.png?raw=true" alt="guardmesh Discovered Applications" width="800" style="border-radius: 10px;" />
</p>

### Adaptive Mobile Interface
<p align="center">
  <img src="docs/images/guardmesh_mobile.png?raw=true" alt="guardmesh Mobile View" width="370" style="border-radius: 16px; margin-right: 12px;" />
  <img src="docs/images/guardmesh_mobile_menu.png?raw=true" alt="guardmesh Mobile Menu" width="370" style="border-radius: 16px;" />
</p>

---

## Quick Start

### 1. Installation

Build from source (requires Go 1.24+ and Node.js 20+):

```bash
git clone https://github.com/alexandrmotologa/guardmesh.git
cd guardmesh
make build-web
make build
```

Or run directly with Docker:

```bash
docker compose -f examples/docker-compose.yml up -d
```

### 2. Start the Gateway

```bash
./guardmesh start --config ./guardmesh.yaml --port 8080
```

The gateway listens on port `8080`. Access the embedded compliance cockpit at `http://localhost:8080`.

### 3. Route Traffic Through the Proxy

#### Python (OpenAI SDK)

```python
from openai import OpenAI

# Point client to guardmesh proxy with zero code changes to logic
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="your-openai-api-key",
    default_headers={"X-Guardmesh-App-ID": "crm-agent"}
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "user", "content": "Process payment for Alice Walker, Romanian CNP 1800514123459, IBAN DE89370400440532013000."}
    ],
    stream=True
)

for chunk in response:
    content = chunk.choices[0].delta.content
    if content:
        print(content, end="", flush=True)
```

External OpenAI endpoints receive only:
```
"Process payment for {{GUARDMESH_ENTITY_PERSON_NAME_1}}, Romanian CNP {{GUARDMESH_ENTITY_ROMANIAN_CNP_1}}, IBAN {{GUARDMESH_ENTITY_IBAN_1}}."
```

The client receives the response with original values restored.

#### TypeScript (Anthropic SDK)

```typescript
import Anthropic from '@anthropic-ai/sdk';

const anthropic = new Anthropic({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-anthropic-api-key',
  defaultHeaders: { 'X-Guardmesh-App-ID': 'support-bot' },
});

const message = await anthropic.messages.create({
  model: 'claude-3-5-sonnet-20241022',
  max_tokens: 1024,
  messages: [{ role: 'user', content: 'Customer contact is alice@company.org, NIF 12345678Z.' }],
});
```

---

## EU AI Act Risk Tiering

`guardmesh` continuously parses context, system prompts, and tool schemas:

| Tier | Regulation Article | Scope & Triggers | Gateway Action |
|---|---|---|---|
| **Prohibited** | Article 5 | Workplace emotion recognition, social scoring, biometric categorization of protected traits, subliminal manipulation. | Terminated before egress (HTTP 403 Forbidden). Webhook alert sent. |
| **High-Risk** | Annex III | CV screening, recruitment filtering, credit scoring, health insurance risk evaluation, emergency service triage. | Telemetry logged. Article 14 human oversight flags set. Technical dossier enabled. |
| **Limited Risk** | Article 50 | Customer-facing conversational agents, synthetic media generation. | Injects transparency header and disclaimer notice. PII masking active. |
| **Minimal Risk** | Article 4 | Routine text summarization, coding assistance, translation. | Standard telemetry logging. |

---

## CLI Reference

### Audit a System Prompt

Evaluate system instructions or tool definitions against EU AI Act criteria:

```bash
guardmesh classify --prompt "You are an HR agent. Screen incoming resumes and score job applicants."
```

### Scan Files or Text for Sensitive Entities

```bash
guardmesh scan --text "Verify employee EMP-982314 with Romanian CNP 1800514123459, Spanish NIF 12345678Z, email alice@domain.com."
```

### Generate Annex IV Technical Dossier

Produce an official multi-page compliance dossier in PDF or Markdown:

```bash
guardmesh dossier --app-id crm-agent --output ./EU_AI_Act_Dossier_CRMAgent.pdf
```

### List Discovered Applications

```bash
guardmesh apps list
```

### Run Latency Benchmark

```bash
guardmesh benchmark --requests 1000
```

---

## Technical Architecture

```
                       +-----------------------------+
                       |   Internal Enterprise App   |
                       +-----------------------------+
                                      |
                                      | OpenAI / Anthropic SDK Call
                                      v
                       +-----------------------------+
                       |       guardmesh Proxy       |
                       |       (:8080 Gateway)       |
                       +-----------------------------+
                                      |
                         +------------+------------+
                         |                         |
                         v                         v
         +-------------------------------+ +-------------------------------+
         |    Security & Policy Guard    | |    Bidirectional PII Engine   |
         |  - Article 5 Hard Kill        | |  - Luhn, Mod-97, CNP & NIF    |
         |  - Article 50 Transparency    | |  - Custom Enterprise Regex    |
         |  - Webhook Alert Dispatcher   | |  - Ephemeral AES-256 Vault    |
         |  - Annex III Classifier       | |  - Chunk Window Rehydrator    |
         +-------------------------------+ +-------------------------------+
                         |                         |
                         +------------+------------+
                                      |
                                      | Sanitized Request (Zero Real PII)
                                      v
                       +-----------------------------+
                       |     External Cloud LLM      |
                       +-----------------------------+
                                      |
                                      | SSE Token Stream
                                      v
                       +-----------------------------+
                       |    Stream Window Buffer     |
                       |  - Resolves split tokens    |
                       |  - Injects Art. 50 notices  |
                       |  - Restores original PII    |
                       |  - Writes SQLite audit WAL  |
                       +-----------------------------+
                                      |
                                      | Rehydrated Response
                                      v
                       +-----------------------------+
                       |   Internal Enterprise App   |
                       +-----------------------------+
```

---

## Observability & Endpoints

- `GET /healthz` - Kubernetes liveness check
- `GET /readyz` - Readiness probe asserting database and vault health
- `GET /metrics` - Prometheus metrics format
- `GET /api/events/export?format=csv` - Raw audit logs in CSV format
- `GET /api/events/export?format=jsonl` - Raw audit logs in JSON Lines format

---

## Configuration (`guardmesh.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  upstream_openai: "https://api.openai.com"
  upstream_anthropic: "https://api.anthropic.com"
  db_path: "./guardmesh.db"

pii:
  enabled: true
  mask_mode: "token"
  token_ttl_seconds: 300
  detectors:
    credit_card: true
    iban: true
    email: true
    phone: true
    ssn: true
    ip_address: true
    person_name: true
    api_secrets: true
    romanian_cnp: true
    spanish_nif: true
    custom_rules:
      - name: "EmployeeBadge"
        pattern: "\\bEMP-[0-9]{6}\\b"
        category: "EMPLOYEE_ID"

governance:
  enabled: true
  enforce_article5_block: true
  log_annex_iii_high_risk: true
  require_human_oversight_flag: true
  default_application_id: "crm-agent"
  article50_transparency: true
  article50_header: true
  article50_disclaimer_text: "\n\n*(Notice: Response generated by an artificial intelligence system in accordance with Article 50 of Regulation (EU) 2024/1689)*"

security:
  scan_prompt_injection: true
  block_prompt_injection: false

alerts:
  enabled: false
  webhooks:
    - name: "SlackAlerts"
      url: "https://hooks.slack.com/services/T00/B00/XXXX"
      secret: "whsec_optional_hmac_secret"
      events: ["blocked", "prohibited", "high_risk"]
```

---

## License

Apache 2.0. See [LICENSE](LICENSE) for details.
