# guardmesh Architecture Specification

`guardmesh` operates as an in-line reverse proxy situated between internal enterprise applications and external large language model providers. It enforces data privacy and regulatory compliance before any request payload leaves the corporate network boundary.

```
+---------------------------+
| Internal Application      | (CRM, Internal Agent, Helpdesk, Slack Bot)
+---------------------------+
              |
              | OpenAI or Anthropic SDK HTTP request
              v
+---------------------------+
| guardmesh Gateway (:8080) |
+---------------------------+
              |
      +-------+-------------------------+
      |                                 |
      v                                 v
+---------------------------+   +---------------------------+
| EU AI Act Classifier      |   | PII De-Identification     |
| - Article 5 check         |   | - Regex & Luhn & Mod-97   |
| - Annex III check         |   | - In-memory AES-GCM vault |
| - Prompt injection scan   |   | - Deterministic tokens    |
+---------------------------+   +---------------------------+
      |                                 |
      +-------+-------------------------+
              |
              | Sanitized request (Zero real PII)
              v
+---------------------------+
| External LLM Endpoint     | (api.openai.com, api.anthropic.com)
+---------------------------+
              |
              | SSE token chunks or JSON response
              v
+---------------------------+
| Stream Window Rehydrator  |
| - Chunk buffer lookahead  |
| - Decrypts tokens via TTL |
| - Restores original PII   |
+---------------------------+
              |
              | Re-hydrated response
              v
+---------------------------+
| Internal Application      |
+---------------------------+
```

## 1. Request Interception Pipeline

When a client application initiates an OpenAI or Anthropic API request, `guardmesh` processes it through four stages:

1. **Context Extraction**: The gateway parses the request body and extracts system instructions, user messages, assistant context, and tool specifications.
2. **Regulatory Risk Evaluation**: The text context is evaluated against predefined legal heuristics mapped directly to Regulation (EU) 2024/1689. If a prohibited AI practice (Article 5) is detected and enforcement is enabled, the request is terminated immediately with an HTTP 403 Forbidden status code.
3. **Sensitive Entity De-Identification**: The PII detector runs compiled regular expressions and algorithmic validations (such as Luhn for payment cards and Mod-97 for IBANs). Detected sensitive values are placed in an ephemeral in-memory vault and substituted with deterministic placeholders such as `{{GUARDMESH_ENTITY_IBAN_1}}`.
4. **Upstream Forwarding**: The sanitized JSON payload is forwarded to the configured upstream endpoint with original client authentication headers preserved.

## 2. Cryptographic Token Vault

The token vault operates entirely in volatile memory and does not write unmasked personal data to disk:

- **Symmetric Encryption**: Entity values inside the vault are encrypted using AES-256 GCM. The encryption key is randomly generated upon process startup and never leaves RAM.
- **Session Isolation**: Each request or session maintains an isolated token namespace. Values mapped in one request cannot collide with or resolve tokens from another request.
- **Time-to-Live Expiry**: Mappings are held with a configurable TTL (default 300 seconds). A background sweeper purges expired records to prevent memory exhaustion.

## 3. Streaming Token Re-Hydration Across Chunk Boundaries

Cloud model providers deliver responses via Server-Sent Events (SSE). Upstream models frequently divide words or special tokens across arbitrary chunk boundaries.

A standard string replacement fails if `{{GUARDMESH_ENT` arrives in chunk 1 and `ITY_PERSON_1}}` arrives in chunk 2.

To solve this, `guardmesh` uses a sliding lookahead window:
- When an opening delimiter `{{` is encountered without a matching `}}`, the trailing bytes are retained in the buffer.
- When the subsequent SSE chunk arrives, the buffer is concatenated and evaluated.
- Once the complete token is formed, the original value is decrypted from the session vault and emitted in the SSE delta stream.
- If the buffer exceeds 64 bytes without finding a closing delimiter, the pending text is emitted as standard content.

## 4. SQLite Audit Store in WAL Mode

Operational telemetry is written to a local SQLite database configured with:
- `PRAGMA journal_mode = WAL;` (Write-Ahead Logging allows concurrent readers and writers without lock contention).
- `PRAGMA synchronous = NORMAL;` (Reduces disk sync calls while maintaining integrity).
- `PRAGMA busy_timeout = 5000;` (Prevents transient lock errors under load).

Audit events capture request identifiers, timestamps, model names, latency in milliseconds, token counts, risk classification tiers, detected PII counts, and SHA-256 prompt fingerprints. Real prompt text is never stored in full; only sanitized snippets and cryptographic hashes are persisted.
