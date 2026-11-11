# guardmesh

Enterprise AI governance gateway, bidirectional PII de-identification proxy, and automated EU AI Act compliance auditor.

Deploying generative AI within corporate infrastructure introduces regulatory liabilities under **Regulation (EU) 2024/1689 (EU AI Act)** and **GDPR**. Violations of prohibited practices carry statutory fines up to €35,000,000 or 7% of global annual turnover. Internal microservices routinely transmit customer personal data, developer secrets, and internal documents to third-party model providers.

`guardmesh` acts as an on-premises reverse proxy between internal applications and external model providers (OpenAI, Anthropic). It strips personal data and secrets before outbound transmission, reconstructs original values in streaming responses, classifies workloads into EU AI Act risk tiers, blocks prohibited practices, and generates regulator-ready technical documentation dossiers in PDF and Markdown.

---

## Key Capabilities

- **Zero-Code SDK Integration**: Compatible with OpenAI and Anthropic API endpoints. Redirect calls by updating the base URL (for example, `http://localhost:8080/v1`).
- **Bidirectional PII De-Identification**: Scans for credit cards (Luhn verified), bank accounts (IBAN mod-97 verified), emails, phone numbers, SSNs, IP addresses, person names, and developer API keys. Replaces sensitive entities with deterministic tokens (`{{GUARDMESH_ENTITY_IBAN_1}}`).
- **Streaming Token Boundary Buffer**: Resolves tokens split across Server-Sent Events (SSE) packet boundaries without increasing stream latency.
- **Article 5 Hard-Kill Enforcer**: Rejects prohibited AI practices (such as workplace emotion recognition, social scoring, or biometric categorization) with an HTTP 403 Forbidden status code.
- **Annex III High-Risk Classification**: Evaluates system instructions and tool definitions against Annex III domains (recruitment, credit assessment, essential public services, education).
- **1-Click Audit Dossier Generation**: Exports complete Annex IV technical documentation and NIST AI RMF cross-walks in PDF and Markdown format.
- **Embedded Compliance Cockpit**: Serves a real-time React 19 dashboard directly from the single Go binary at `http://localhost:8080/`.
- **Sub-Millisecond Overhead**: Operates with an average proxy overhead of under 0.1 milliseconds per request.

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
        {"role": "user", "content": "Process payment for John Doe at DE89370400440532013000, SSN 123-45-6789."}
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
"Process payment for {{GUARDMESH_ENTITY_PERSON_NAME_1}} at {{GUARDMESH_ENTITY_IBAN_1}}, SSN {{GUARDMESH_ENTITY_SSN_1}}."
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
  messages: [{ role: 'user', content: 'Customer email is alice@company.org.' }],
});
```

---

## EU AI Act Risk Tiering

`guardmesh` continuously parses context, system prompts, and tool schemas:

| Tier | Regulation Article | Scope & Triggers | Gateway Action |
|---|---|---|---|
| **Prohibited** | Article 5 | Workplace emotion recognition, social scoring, biometric categorization of protected traits, subliminal manipulation. | Terminated before egress (HTTP 403 Forbidden). Audit record logged. |
| **High-Risk** | Annex III | CV screening, recruitment filtering, credit scoring, health insurance risk evaluation, emergency service triage. | Telemetry logged. Article 14 human oversight flags set. Technical dossier enabled. |
| **Limited Risk** | Article 50 | Customer-facing conversational agents, synthetic media generation. | Transparency verification. PII masking active. |
| **Minimal Risk** | Article 4 | Routine text summarization, coding assistance, translation. | Standard telemetry logging. |

---

## CLI Reference

### Audit a System Prompt

Evaluate system instructions or tool definitions against EU AI Act criteria:

```bash
guardmesh classify --prompt "You are an HR agent. Screen incoming resumes and score job applicants."
```

Output:
```
------------------------------------------------------------------
 EU AI Act Compliance Evaluation (Regulation EU 2024/1689)
------------------------------------------------------------------
 Risk Classification Tier: HIGH_RISK
 Risk Severity Score:     0.85 / 1.00
 Automated Hard Block:    false

 Applicable Legal Articles:
  • Annex III(4)(a)

 Classification Reasons & Findings:
  • Recruitment, CV screening, job application filtering or candidate evaluation

 Required Compliance Controls:
  [X] Article 14 Human Oversight, Article 11 Technical Documentation

 ✓ Security Check: No adversarial prompt injection detected
------------------------------------------------------------------
```

### Scan Files or Text for Sensitive Entities

```bash
guardmesh scan --text "Customer Alice Walker at DE89370400440532013000, SSN 123-45-6789, email alice@domain.com."
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
         |  - Article 5 Hard Kill        | |  - Luhn & Mod-97 Checksums    |
         |  - Annex III Classifier       | |  - Ephemeral AES-256 Vault    |
         |  - Injection Scanner          | |  - Chunk Window Rehydrator    |
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

### Zero-Allocation Entity Detection
- **Payment Cards**: Visa, MasterCard, Amex, and Discover patterns validated against the Luhn algorithm to prevent false positives on numerical sequences.
- **Bank Accounts**: IBAN detection with country-specific length checks and ISO 7064 Mod-97-10 verification.
- **Developer Secrets**: Scans for OpenAI API keys (`sk-...`), GitHub personal access tokens (`ghp_...`), AWS credentials (`AKIA...`), JWT tokens, and private RSA/EC keys.
- **Identifiers**: US Social Security Numbers, international phone numbers (E.164), non-loopback IP addresses, and person names.

### Cryptographic Token Vault
Mapped values reside in volatile RAM under an AES-256 GCM cipher. The key is generated at startup and held only in memory. Mappings automatically purge based on a configurable time-to-live setting (default 300 seconds).

---

## Performance

Tested on a standard 4-core x86_64 system processing realistic conversational payloads containing multiple entity types (names, emails, IBANs, SSNs, secrets):

| Operation | Metric |
|---|---|
| PII Detection + Tokenization | 38.4 µs |
| In-Memory AES-GCM Decryption | 12.1 µs |
| Stream Chunk Window Lookahead | 8.2 µs |
| Total Gateway Overhead | **0.068 ms** |
| Throughput | **14,700+ ops/sec** |
| Idle Binary Memory | **18 MB** |

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

governance:
  enabled: true
  enforce_article5_block: true
  log_annex_iii_high_risk: true
  require_human_oversight_flag: true
  default_application_id: "crm-agent"

security:
  scan_prompt_injection: true
  block_prompt_injection: false
```

---

## License

Apache 2.0. See [LICENSE](LICENSE) for details.
