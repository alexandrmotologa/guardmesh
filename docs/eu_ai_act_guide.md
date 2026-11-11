# EU AI Act Compliance Guide (Regulation EU 2024/1689)

The European Union Artificial Intelligence Act entered into force on 1 August 2024. It creates a risk-based regulatory framework that applies to providers and deployers of AI systems within the European Union or where system outputs are used in the EU.

This guide details how `guardmesh` automates compliance obligations across risk tiers.

---

## 1. Classification Matrix Overview

| Tier | Legal Basis | Scope & Application | Enforcement in guardmesh |
|---|---|---|---|
| **Prohibited** | Article 5 | Practices with unacceptable risk to human dignity, safety, or fundamental rights | Automated hard rejection with HTTP 403 status code and security audit record |
| **High-Risk** | Article 6 & Annex III | Critical areas: recruitment, credit scoring, education, essential services | Mandatory telemetry logging, Article 14 human oversight flags, technical dossier generation |
| **Limited Risk** | Article 50 | Conversational AI and synthetic content | Transparency checks and PII de-identification |
| **Minimal Risk** | Article 4 | Routine business tools, coding assistants, spam filters | Baseline performance tracking and monitoring |

---

## 2. Prohibited AI Practices (Article 5)

Under Article 5, specific AI applications are banned across all EU member states. `guardmesh` scans incoming context against these criteria:

- **Article 5(1)(a)**: Subliminal techniques beyond a person's consciousness or purposefully manipulative methods that distort behavior.
- **Article 5(1)(c)**: Social scoring leading to detrimental or unfavorable treatment of natural persons.
- **Article 5(1)(d)**: Predictive policing based solely on profiling or personality traits without objective facts.
- **Article 5(1)(e)**: Untargeted scraping of facial images from the internet or CCTV footage to build facial recognition databases.
- **Article 5(1)(f)**: Emotion recognition in workplaces or educational institutions, unless for medical or safety reasons.
- **Article 5(1)(g)**: Biometric categorization to deduce protected attributes such as race, political opinions, trade union membership, religious beliefs, or sexual orientation.

When a request matches any Article 5 pattern and `enforce_article5_block` is enabled, `guardmesh` terminates the request before sending data upstream and returns an error response:

```json
{
  "error": {
    "code": "eu_ai_act_prohibited",
    "message": "Blocked under EU AI Act Article 5: Prohibited AI Practice (Article 5(1)(f))",
    "risk_tier": "PROHIBITED",
    "applicable_articles": ["Article 5(1)(f)"],
    "reasons": ["Emotion recognition in workplace or educational institutions"]
  }
}
```

---

## 3. High-Risk AI Systems (Annex III)

Systems categorized as High-Risk must comply with mandatory requirements before placement on the market or putting into service:

### Annex III Domains Monitored by guardmesh
- **Annex III(4)(a)**: Recruitment, CV filtering, applicant evaluation, and candidate screening.
- **Annex III(4)(b)**: Task allocation based on worker behavior, productivity surveillance, promotion, and termination decisions.
- **Annex III(5)(b)**: Creditworthiness assessment and credit scoring of natural persons.
- **Annex III(5)(c)**: Risk assessment and pricing for life and health insurance.
- **Annex III(5)(d)**: Emergency service call prioritization and dispatch triage.
- **Annex III(3)(a)-(b)**: Admissions assessments, exam grading, and student monitoring.

### Mandatory Technical Measures
1. **Article 11 (Technical Documentation)**: Deployers must maintain current technical documentation pursuant to Annex IV. `guardmesh` generates this documentation via CLI (`guardmesh dossier`) or through the embedded cockpit.
2. **Article 14 (Human Oversight)**: High-risk systems must enable human operators to understand outputs, verify decisions, and intervene or stop system operation.
3. **Article 15 (Accuracy, Robustness and Cybersecurity)**: Systems must withstand adversarial attacks such as prompt injection and data poisoning.

---

## 4. Technical Documentation Dossier (Annex IV)

Under Article 11, technical documentation must contain:
1. General description of the AI system, intended purpose, and version history.
2. Model details, specifications, and upstream integrations.
3. Risk management system results and risk evaluations.
4. Data governance records, personal data categories, and de-identification measures.
5. Human oversight controls, override capabilities, and operational limits.
6. Cybersecurity defenses and resilience verification.
7. Audit log samples and cryptographic verification fingerprints.

`guardmesh` generates this documentation directly in PDF and Markdown format for regulatory review.
