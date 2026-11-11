package pii

import (
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/config"
)

func TestDetectorAndVault(t *testing.T) {
	cfg := config.DefaultConfig()
	det := NewDetector(cfg.PII.Detectors)
	vault, err := NewVault(5 * time.Minute)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}
	defer vault.Close()

	// Text with multiple PII elements
	text := "Customer John Doe with email john.doe@example.com, SSN 123-45-6789, IP 198.51.100.42, and key sk-1234567890abcdef1234567890."

	masked, count, cats := vault.MaskText("test-sess-1", text, det, "token")
	if count < 4 {
		t.Fatalf("expected at least 4 entities detected, got %d", count)
	}
	if len(cats) < 3 {
		t.Fatalf("expected at least 3 categories, got %v", cats)
	}

	// Verify original values are not in masked text
	if strings.Contains(masked, "john.doe@example.com") {
		t.Errorf("masked text still contains email")
	}
	if strings.Contains(masked, "123-45-6789") {
		t.Errorf("masked text still contains SSN")
	}
	if strings.Contains(masked, "sk-1234567890abcdef1234567890") {
		t.Errorf("masked text still contains API secret")
	}

	// Rehydrate text
	rehydrated := vault.RehydrateText("test-sess-1", masked)
	if !strings.Contains(rehydrated, "john.doe@example.com") {
		t.Errorf("rehydrated text missing email: %s", rehydrated)
	}
	if !strings.Contains(rehydrated, "123-45-6789") {
		t.Errorf("rehydrated text missing SSN: %s", rehydrated)
	}
}

func TestStreamRehydrator_SplitChunks(t *testing.T) {
	cfg := config.DefaultConfig()
	det := NewDetector(cfg.PII.Detectors)
	vault, err := NewVault(5 * time.Minute)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}
	defer vault.Close()

	original := "Contact support at alice@company.org immediately."
	masked, count, _ := vault.MaskText("stream-sess", original, det, "token")
	if count == 0 {
		t.Fatalf("expected email detected")
	}

	// Simulate streaming token split:
	// Find the token in masked text
	openIdx := strings.Index(masked, "{{")
	closeIdx := strings.Index(masked, "}}")
	if openIdx == -1 || closeIdx == -1 {
		t.Fatalf("token delimiters not found in: %s", masked)
	}

	token := masked[openIdx : closeIdx+2]
	splitPoint := len(token) / 2
	tokenPart1 := token[:splitPoint]
	tokenPart2 := token[splitPoint:]

	chunk1 := masked[:openIdx] + tokenPart1
	chunk2 := tokenPart2 + masked[closeIdx+2:]

	rehydrator := NewStreamRehydrator("stream-sess", vault)

	out1 := rehydrator.ProcessChunk(chunk1)
	out2 := rehydrator.ProcessChunk(chunk2)
	outFinal := out1 + out2 + rehydrator.FlushRemainder()

	if !strings.Contains(outFinal, "alice@company.org") {
		t.Fatalf("expected stream rehydrator to reconstruct original email across split chunks, got: %q", outFinal)
	}
	if strings.Contains(outFinal, "{{") || strings.Contains(outFinal, "}}") {
		t.Fatalf("stream rehydrator left unhydrated token marks: %q", outFinal)
	}
}
