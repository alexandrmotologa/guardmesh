#!/usr/bin/env python3
"""
Seed guardmesh SQLite database with realistic enterprise governance telemetry.
"""

import sqlite3
import datetime
import uuid
import hashlib

DB_PATH = "guardmesh.db"

conn = sqlite3.connect(DB_PATH)
cur = conn.cursor()

# Ensure schema exists
cur.execute("""
CREATE TABLE IF NOT EXISTS audit_events (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    app_id TEXT NOT NULL,
    client_ip TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    latency_ms REAL NOT NULL,
    prompt_tokens INTEGER NOT NULL,
    completion_tokens INTEGER NOT NULL,
    risk_tier TEXT NOT NULL,
    risk_score REAL NOT NULL,
    risk_reasons TEXT NOT NULL,
    is_blocked INTEGER NOT NULL,
    blocked_reason TEXT,
    pii_detected_count INTEGER NOT NULL,
    pii_categories TEXT NOT NULL,
    prompt_sha256 TEXT NOT NULL,
    sample_snippet TEXT NOT NULL
);
""")

cur.execute("""
CREATE TABLE IF NOT EXISTS applications (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    highest_risk_tier TEXT NOT NULL,
    total_requests INTEGER NOT NULL DEFAULT 0,
    total_pii_masked INTEGER NOT NULL DEFAULT 0,
    compliance_status TEXT NOT NULL
);
""")

# Clear previous data
cur.execute("DELETE FROM audit_events;")
cur.execute("DELETE FROM applications;")

apps_data = [
    ("crm-financial-assistant", "Enterprise CRM & Billing Gateway", "HIGH_RISK", 1480, 4210, "COMPLIANT"),
    ("hr-recruiting-engine", "Talent Acquisition & CV Parser", "HIGH_RISK", 620, 1840, "NEEDS_REVIEW"),
    ("customer-support-bot", "Customer Service Live Chat", "LIMITED_RISK", 3920, 5630, "COMPLIANT"),
    ("internal-developer-copilot", "Engineering Architecture & Code Assistant", "MINIMAL_RISK", 5410, 890, "COMPLIANT"),
    ("biometric-attendance-tracker", "Biometric Workplace Surveillance Scanner", "PROHIBITED", 48, 0, "VIOLATION"),
    ("credit-scoring-agent", "Retail Banking Creditworthiness Assessor", "HIGH_RISK", 310, 840, "COMPLIANT"),
]

now = datetime.datetime.now(datetime.timezone.utc)

for app_id, name, tier, reqs, pii, status in apps_data:
    first_seen = (now - datetime.timedelta(days=14)).isoformat()
    last_seen = (now - datetime.timedelta(minutes=5)).isoformat()
    cur.execute("""
        INSERT INTO applications (id, name, first_seen, last_seen, highest_risk_tier, total_requests, total_pii_masked, compliance_status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?);
    """, (app_id, name, first_seen, last_seen, tier, reqs, pii, status))

# Seed realistic audit events
events = [
    # 1. Prohibited Article 5 hard blocks
    {
        "app_id": "biometric-attendance-tracker",
        "provider": "openai",
        "model": "gpt-4o",
        "status_code": 403,
        "latency_ms": 0.8,
        "prompt_tokens": 120,
        "completion_tokens": 0,
        "risk_tier": "PROHIBITED",
        "risk_score": 1.0,
        "risk_reasons": '["Workplace emotion recognition prohibited by Article 5(1)(f)"]',
        "is_blocked": 1,
        "blocked_reason": "EU AI Act Article 5 Prohibited AI Practice (Emotion Recognition in Workplace)",
        "pii_detected_count": 0,
        "pii_categories": "[]",
        "sample": "Analyze employee video feed during shift to evaluate micro-expressions, fatigue, and frustration...",
    },
    {
        "app_id": "hr-recruiting-engine",
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022",
        "status_code": 403,
        "latency_ms": 0.9,
        "prompt_tokens": 145,
        "completion_tokens": 0,
        "risk_tier": "PROHIBITED",
        "risk_score": 0.95,
        "risk_reasons": '["Biometric categorization of sensitive protected traits Article 5(1)(c)"]',
        "is_blocked": 1,
        "blocked_reason": "Article 5 Prohibited: Inferring political or philosophical beliefs from candidate facial geometry",
        "pii_detected_count": 2,
        "pii_categories": '["PERSON_NAME","EMAIL"]',
        "sample": "Classify candidate biometric imagery to infer religious affiliation and social standing...",
    },
    # 2. High-Risk Annex III requests (Passed with PII sanitization)
    {
        "app_id": "crm-financial-assistant",
        "provider": "openai",
        "model": "gpt-4o",
        "status_code": 200,
        "latency_ms": 1.2,
        "prompt_tokens": 280,
        "completion_tokens": 190,
        "risk_tier": "HIGH_RISK",
        "risk_score": 0.85,
        "risk_reasons": '["Annex III Section 5(b): Creditworthiness assessment & financial risk scoring"]',
        "is_blocked": 0,
        "blocked_reason": None,
        "pii_detected_count": 4,
        "pii_categories": '["ROMANIAN_CNP","IBAN","CREDIT_CARD","EMAIL"]',
        "sample": "Process high-volume loan application for borrower Elena Popescu, CNP 2890514123459, IBAN RO49AAAA...",
    },
    {
        "app_id": "hr-recruiting-engine",
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022",
        "status_code": 200,
        "latency_ms": 1.4,
        "prompt_tokens": 450,
        "completion_tokens": 210,
        "risk_tier": "HIGH_RISK",
        "risk_score": 0.80,
        "risk_reasons": '["Annex III Section 4(a): Recruitment and candidate resume evaluation"]',
        "is_blocked": 0,
        "blocked_reason": None,
        "pii_detected_count": 3,
        "pii_categories": '["PERSON_NAME","PHONE","SSN"]',
        "sample": "Extract engineering competencies from candidate CV for Michael Chen, phone +1-415-555-0199, SSN...",
    },
    {
        "app_id": "credit-scoring-agent",
        "provider": "openai",
        "model": "gpt-4o",
        "status_code": 200,
        "latency_ms": 1.1,
        "prompt_tokens": 310,
        "completion_tokens": 140,
        "risk_tier": "HIGH_RISK",
        "risk_score": 0.82,
        "risk_reasons": '["Annex III Section 5(b): Risk evaluation for commercial credit limits"]',
        "is_blocked": 0,
        "blocked_reason": None,
        "pii_detected_count": 3,
        "pii_categories": '["SPANISH_NIF","IBAN","PERSON_NAME"]',
        "sample": "Evaluate credit line extension for Carlos Garcia, Spanish NIF 12345678Z, account ES9121000418450200051332...",
    },
    # 3. Limited Risk Article 50 requests
    {
        "app_id": "customer-support-bot",
        "provider": "openai",
        "model": "gpt-4o-mini",
        "status_code": 200,
        "latency_ms": 0.6,
        "prompt_tokens": 160,
        "completion_tokens": 95,
        "risk_tier": "LIMITED_RISK",
        "risk_score": 0.40,
        "risk_reasons": '["Article 50: Interactive conversational AI chatbot requires transparency notice"]',
        "is_blocked": 0,
        "blocked_reason": None,
        "pii_detected_count": 2,
        "pii_categories": '["EMAIL","PHONE"]',
        "sample": "Help customer resolve return request for order #8492. User email support@company.org...",
    },
    # 4. Minimal Risk requests
    {
        "app_id": "internal-developer-copilot",
        "provider": "anthropic",
        "model": "claude-3-5-sonnet-20241022",
        "status_code": 200,
        "latency_ms": 0.7,
        "prompt_tokens": 620,
        "completion_tokens": 340,
        "risk_tier": "MINIMAL_RISK",
        "risk_score": 0.10,
        "risk_reasons": '["Routine software engineering code synthesis and refactoring"]',
        "is_blocked": 0,
        "blocked_reason": None,
        "pii_detected_count": 1,
        "pii_categories": '["API_SECRET"]',
        "sample": "Optimize Go goroutine worker pool with bounded channels and mutex lock synchronization...",
    },
]

# Generate 30 events spread over the last few hours
import random

for i in range(35):
    base = random.choice(events)
    event_id = f"evt_{uuid.uuid4().hex[:12]}"
    delta_mins = random.randint(1, 180)
    ts = (now - datetime.timedelta(minutes=delta_mins)).isoformat()
    ip = f"10.0.{random.randint(1, 5)}.{random.randint(10, 240)}"
    prompt_hash = hashlib.sha256(base["sample"].encode()).hexdigest()
    
    cur.execute("""
        INSERT INTO audit_events (
            id, timestamp, app_id, client_ip, provider, model, status_code,
            latency_ms, prompt_tokens, completion_tokens, risk_tier, risk_score,
            risk_reasons, is_blocked, blocked_reason, pii_detected_count,
            pii_categories, prompt_sha256, sample_snippet
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
    """, (
        event_id, ts, base["app_id"], ip, base["provider"], base["model"],
        base["status_code"], base["latency_ms"] + random.uniform(-0.2, 0.4),
        base["prompt_tokens"], base["completion_tokens"], base["risk_tier"],
        base["risk_score"], base["risk_reasons"], base["is_blocked"],
        base["blocked_reason"], base["pii_detected_count"],
        base["pii_categories"], prompt_hash, base["sample"]
    ))

conn.commit()
conn.close()
print("✓ Seeded guardmesh.db with 6 applications and 35 realistic audit events.")
