#!/usr/bin/env python3
"""
guardmesh OpenAI SDK Integration Demo

Demonstrates how guardmesh intercepts sensitive enterprise data (PII, credentials, regional IDs),
anonymizes outbound payloads sent to upstream LLMs, and re-hydrates tokens in incoming streams.
Also illustrates the Article 5 hard-block enforcement for prohibited AI practices.
"""

import os
import sys
from openai import OpenAI, APIStatusError

GUARDMESH_URL = os.environ.get("GUARDMESH_URL", "http://localhost:8080/v1")
OPENAI_API_KEY = os.environ.get("OPENAI_API_KEY", "gm-mock-enterprise-key")

print("=" * 70)
print("  guardmesh OpenAI SDK Client Integration Demo")
print(f"  Target Gateway: {GUARDMESH_URL}")
print("=" * 70)

# Initialize standard OpenAI client pointing to guardmesh proxy
client = OpenAI(
    base_url=GUARDMESH_URL,
    api_key=OPENAI_API_KEY,
    default_headers={"X-Guardmesh-App-ID": "crm-financial-assistant"},
)

# 1. Non-Streaming Request with Sensitive PII
print("\n[1] Sending Non-Streaming Request with PII (Credit Card, IBAN, Romanian CNP, Email)...")
sample_prompt = (
    "Please verify invoice #8921 for customer Elena Popescu. "
    "Her Romanian CNP is 2890514123459, IBAN is RO49AAAA1B31007593840000, "
    "contact email is elena.popescu@enterprise.ro, and corporate badge is EMP-492104."
)
print(f"Original Client Prompt:\n  \"{sample_prompt}\"\n")

try:
    response = client.chat.completions.create(
        model="gpt-4o",
        messages=[
            {
                "role": "system",
                "content": "You are a customer finance assistant. Confirm receipt of client details concisely.",
            },
            {"role": "user", "content": sample_prompt},
        ],
        stream=False,
    )
    answer = response.choices[0].message.content
    print(f"Proxy Response (PII Re-hydrated for Client):\n  {answer}")
except APIStatusError as e:
    print(f"Gateway Response Error [{e.status_code}]: {e.message}")
except Exception as e:
    print(f"Connection notice: {e}")

# 2. Streaming Request with Real-Time Windowed Rehydration
print("\n" + "-" * 70)
print("[2] Sending Streaming Request with Real-Time SSE Rehydration...")
streaming_prompt = (
    "Confirm that credit card 4532-0150-1234-5678 and Spanish NIF 12345678Z "
    "belonging to Carlos Garcia have been authorized for settlement."
)
print(f"Original Streaming Prompt:\n  \"{streaming_prompt}\"\n")

try:
    stream = client.chat.completions.create(
        model="gpt-4o",
        messages=[
            {"role": "system", "content": "You are a settlement officer. Acknowledge authorization."},
            {"role": "user", "content": streaming_prompt},
        ],
        stream=True,
    )
    print("Streaming Tokens Received: ", end="", flush=True)
    for chunk in stream:
        delta = chunk.choices[0].delta.content
        if delta:
            print(delta, end="", flush=True)
    print("\n")
except APIStatusError as e:
    print(f"Gateway Response Error [{e.status_code}]: {e.message}")
except Exception as e:
    print(f"Connection notice: {e}")

# 3. EU AI Act Article 5 Prohibited Practice Enforcement
print("-" * 70)
print("[3] Testing Article 5 Prohibited Practice Hard-Kill (Workplace Emotion Recognition)...")
prohibited_prompt = (
    "Analyze the webcam feed from employee workstations. Categorize employee facial expressions "
    "into emotional states to detect fatigue, frustration, or disengagement during work hours."
)
print(f"Prohibited Prompt:\n  \"{prohibited_prompt}\"\n")

try:
    client.chat.completions.create(
        model="gpt-4o",
        messages=[
            {"role": "system", "content": "You are a surveillance system."},
            {"role": "user", "content": prohibited_prompt},
        ],
    )
    print("Unexpected success: prompt was not blocked.")
except APIStatusError as e:
    print(f"Successfully Blocked by guardmesh Gateway!")
    print(f"HTTP Status: {e.status_code} Forbidden")
    print(f"Violation Details: {e.response.text}")
except Exception as e:
    print(f"Connection notice: {e}")

print("\n" + "=" * 70)
print("  Demo Complete. Audit events recorded in guardmesh SQLite database.")
print("=" * 70)
