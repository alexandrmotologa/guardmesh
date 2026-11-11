# `guardmesh` — Architecture & Engineering Master Plan
> **Enterprise AI Governance Gateway, Real-Time Bidirectional PII Anonymizer & Automated EU AI Act Compliance Auditor**  
> *Target: Enterprise C-Suite, Head of AI Deployment, CISO & Compliance Leads*

---

## 1. Executive Summary & Market Problem

### 1.1 The Regulatory Reckoning for Enterprise AI
With the enforcement of the **EU Artificial Intelligence Act (Regulation EU 2024/1689)** alongside **GDPR, HIPAA, and NIST AI RMF**, deploying generative AI models into corporate environments creates unprecedented legal and financial liabilities:
- **Severe Non-Compliance Penalties**: Violations of prohibited practices carry fines up to **€35,000,000 or 7% of global annual turnover**, while breaches of high-risk obligations reach up to **€15,000,000 or 3%**.
- **Shadow AI and Data Leakage**: Employees and internal microservices routinely transmit customer PII, confidential intellectual property, database credentials, and API keys to third-party cloud LLMs.
- **Mandatory Technical Documentation**: Under Article 11 and Annex IV of the EU AI Act, systems classified as High-Risk or General Purpose AI (GPAI) must maintain living technical documentation, risk management logs, and data governance records before deployment.
- **Compliance Latency**: Traditional manual compliance audits take 3 to 6 months per AI initiative, paralyzing innovation and enterprise adoption.

### 1.2 The `guardmesh` Solution
`guardmesh` is a high-performance, transparent reverse proxy and automated governance daemon written in Go with an embedded React 19 compliance cockpit:
1. **Drop-in OpenAI and Anthropic Gateway**: Redirect any existing SDK call by updating the base URL (e.g. from `api.openai.com/v1` to `localhost:8080/v1`) with zero client code modifications.
2. **Bidirectional Zero-Leak PII De-Identification**: Automatically strips emails, phone numbers, credit cards, IBANs, SSNs, IP addresses, names, and developer API secrets using deterministic replacement tokens (`{{GUARDMESH_ENTITY_IBAN_1}}`), sends sanitized payloads to upstream models, and reconstructs real values in streaming responses.
3. **Chunk-Boundary Stream Windowing**: Guarantees that split tokens across Server-Sent Events (SSE) packet boundaries are seamlessly buffered and re-hydrated without dropping stream latency.
4. **Automated EU AI Act Risk Tiering**: Inspects incoming system prompts, message context, and tool definitions to classify workloads into *Prohibited (Article 5)*, *High-Risk (Annex III)*, *Limited Risk (Article 50)*, or *Minimal Risk*.
5. **Article 5 Hard-Kill Enforcer**: Automatically rejects prohibited AI practices (such as biometric categorization, emotion recognition in workplaces, or social scoring) with an HTTP 403 Forbidden compliance report.
6. **1-Click Audit Dossier Generator**: Generates regulator-ready EU AI Act Annex IV technical documentation and NIST AI RMF cross-walk dossiers in multi-page PDF and Markdown formats.

---

## 2. System Architecture

```
                       +---------------------------------------------+
                       |           Internal Enterprise App           |
                       |    (CRM, Chatbot, Backend Agent, Slack)     |
                       +---------------------------------------------+
                                              |
                                              | (OpenAI / Anthropic SDK Call)
                                              v
                       +---------------------------------------------+
                       |             guardmesh Proxy                 |
                       |       (Single Go Binary, <25MB RAM)         |
                       +---------------------------------------------+
                                              |
                         +--------------------+--------------------+
                         |                                         |
                         v                                         v
         +-------------------------------+         +-------------------------------+
         |     Security & Policy Guard   |         |    Bidirectional PII Engine   |
         |  - Article 5 Prohibited Kill  |         |  - High-Speed Regex & Checks  |
         |  - Annex III Risk Classifier  |         |  - Luhn & IBAN Validation     |
         |  - Secret & Credential Filter |         |  - Ephemeral AES Token Vault  |
         |  - Injection & Jailbreak Scan |         |  - Streaming Boundary Buffer  |
         +-------------------------------+         +-------------------------------+
                         |                                         |
                         +--------------------+--------------------+
                                              |
                                              | (Sanitized / Zero PII Request)
                                              v
                       +---------------------------------------------+
                       |              External Cloud LLM             |
                       |      (OpenAI, Anthropic, Gemini, Mistral)   |
                       +---------------------------------------------+
                                              |
                                              | (Streaming Token Response / SSE)
                                              v
                       +---------------------------------------------+
                       |             guardmesh Proxy                 |
                       |  - Real-time PII Token Re-Hydration         |
                       |  - Latency, Token & Cost Accounting         |
                       |  - Immutable Audit Logging in SQLite (WAL)  |
                       +---------------------------------------------+
                                              |
                                              | (Original PII Restored)
                                              v
                       +---------------------------------------------+
                       |           Internal Enterprise App           |
                       +---------------------------------------------+
                                              |
                                              v
                       +---------------------------------------------+
                       |    Embedded React 19 Compliance Cockpit     |
                       |  - Real-time EU AI Act Risk Distribution    |
                       |  - PII Interception Metrics & Shadow AI Map |
                       |  - Interactive Prompt Compliance Auditor    |
                       |  - 1-Click Annex IV PDF Dossier Export      |
                       +---------------------------------------------+
```

---

## 3. Bidirectional PII Sanitization Flow

To prevent training data ingestion and ensure GDPR / HIPAA compliance, `guardmesh` guarantees external providers never observe unmasked sensitive data:

1. **Inbound Client Request**:
   ```json
   {
     "model": "gpt-4o",
     "messages": [
       {"role": "user", "content": "Please verify transaction of 1,200 EUR for Alice Smith, IBAN DE89370400440532013000, SSN 123-45-6789."}
     ]
   }
   ```
2. **`guardmesh` Token Vault (In-Memory AES-256 GCM with TTL)**:
   - `Alice Smith` -> `{{GUARDMESH_ENTITY_PERSON_1}}`
   - `DE89370400440532013000` -> `{{GUARDMESH_ENTITY_IBAN_1}}`
   - `123-45-6789` -> `{{GUARDMESH_ENTITY_SSN_1}}`
3. **Outbound Payload Sent to Cloud Provider**:
   ```json
   {
     "model": "gpt-4o",
     "messages": [
       {"role": "user", "content": "Please verify transaction of 1,200 EUR for {{GUARDMESH_ENTITY_PERSON_1}}, IBAN {{GUARDMESH_ENTITY_IBAN_1}}, SSN {{GUARDMESH_ENTITY_SSN_1}}."}
     ]
   }
   ```
4. **Cloud Provider SSE Stream**:
   - Chunk 1: `{"delta":{"content":"Transaction for {{GUARDMESH_ENT"}}` (buffered in window)
   - Chunk 2: `{"delta":{"content":"ITY_PERSON_1}} confirmed."}}` (resolved to "Alice Smith")
5. **Reconstructed Client Response**:
   `"Transaction for Alice Smith confirmed."`

---

## 4. EU AI Act Classification Matrix

| Tier | Category | Triggers & Examples | Automated Governance Action |
|---|---|---|---|
| **Tier 0** | **Prohibited** (Article 5) | Biometric categorization of sensitive traits, workplace emotion recognition, social scoring, cognitive manipulation. | **Hard Block (HTTP 403)**, security alert, audit log entry. |
| **Tier 1** | **High-Risk** (Annex III) | Employment screening, CV filtering, creditworthiness assessment, critical infrastructure management, education admissions. | Mandatory Annex IV logging, human-in-the-loop flags, risk dossier compilation. |
| **Tier 2** | **Limited Risk** (Article 50) | Customer-facing chatbots, synthetic media generation, deepfake synthesis. | Mandatory transparency disclosures, PII masking. |
| **Tier 3** | **Minimal Risk** | Code synthesis, syntax correction, language translation, document summarization. | Standard latency, cost, and telemetry tracking. |

---

## 5. Repository Structure

```
guardmesh/
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── cmd/
│   └── guardmesh/
│       └── main.go                 # CLI entrypoint (start, classify, scan, dossier, apps, benchmark)
├── internal/
│   ├── config/                     # Configuration loader & defaults
│   │   ├── config.go
│   │   └── rules.go
│   ├── proxy/                      # Streaming Reverse Proxy
│   │   ├── server.go               # HTTP proxy handler & routing
│   │   ├── openai_adapter.go       # OpenAI API compatibility & body rewriting
│   │   ├── anthropic_adapter.go    # Anthropic Claude API compatibility
│   │   └── stream_intercept.go     # SSE sliding window token re-hydration
│   ├── pii/                        # Real-Time De-Identification Engine
│   │   ├── detector.go             # Regex, Luhn, IBAN, Secret & Entity scanners
│   │   ├── vault.go                # Cryptographic ephemeral token map with TTL
│   │   └── rehydrator.go           # Fast string and stream replacer
│   ├── classifier/                 # EU AI Act Compliance Engine
│   │   ├── risk_evaluator.go       # Annex III and Article 5 heuristic evaluator
│   │   ├── policy_rules.go         # Policy rules engine & block actions
│   │   └── injection_guard.go      # Prompt injection & jailbreak scanner
│   ├── dossier/                    # Audit Report Generator
│   │   ├── annex_iv.go             # EU AI Act Technical Documentation builder
│   │   ├── nist_rmf.go             # NIST AI Risk Management Framework exporter
│   │   └── pdf_renderer.go         # Pure Go enterprise PDF generator
│   └── store/                      # SQLite Persistence Layer
│       ├── db.go                   # Schema migrations & WAL configuration
│       ├── audit_events.go         # Event ingestion & query methods
│       └── app_registry.go         # Application tracking & inventory
├── web/                            # Embedded React 19 Compliance Cockpit
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx
│       ├── components/             # RiskHeatmap, PIIStatsCard, LiveEventStream, DossierExport
│       └── pages/                  # Overview, LiveLogs, Applications, Auditor, Settings
├── docs/                           # Architecture, EU AI Act guide, API reference
├── examples/                       # Sample configs, docker-compose, demo scripts
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 6. Implementation Phases

- **Phase 1: Project Scaffolding & Git Lifecycle**
  - Public GitHub repository configuration with verified metadata and topics.
  - Base Go module, Makefile, configuration schemas, and GitHub Actions CI.
- **Phase 2: Bidirectional PII Sanitization & Stream Rehydration**
  - Zero-allocation PII recognizers (Credit Cards with Luhn, IBAN, Emails, Phones, SSNs, IPs, Names, API Keys).
  - AES-256 GCM ephemeral in-memory token vault with TTL expiry.
  - Streaming window buffer resolving split tokens across SSE chunks.
- **Phase 3: High-Speed Reverse Proxy & Multi-Provider Adapters**
  - Transparent OpenAI (`/v1/chat/completions`) and Anthropic (`/v1/messages`) reverse proxy.
  - SQLite WAL audit logger recording latency, tokens, risk scores, and de-identification telemetry.
- **Phase 4: EU AI Act Risk Classifier & Policy Guard**
  - Heuristic classifier evaluating system prompts and tool schemas against Article 5 and Annex III criteria.
  - Article 5 hard-block policy engine (HTTP 403 Forbidden with structured violation details).
  - Prompt injection scanner detecting jailbreak attempts.
- **Phase 5: Automated Regulatory Dossier Generator**
  - Pure Go PDF generator producing multi-page EU AI Act Annex IV technical documentation.
  - NIST AI RMF cross-walk compiler and Markdown export.
- **Phase 6: Embedded React 19 Executive Cockpit**
  - High-performance web UI with live event telemetry, risk distribution charts, shadow AI inventory, and prompt auditor.
  - Single binary distribution using Go `embed`.
- **Phase 7: End-to-End Verification & Benchmark Suite**
  - Comprehensive unit tests, integration tests, mock cloud servers, and CLI test verification.
