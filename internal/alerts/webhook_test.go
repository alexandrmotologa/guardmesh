package alerts

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/store"
)

func TestWebhookDispatcher(t *testing.T) {
	receivedPayload := make(chan *WebhookPayload, 1)
	receivedSig := make(chan string, 1)

	secret := "test-secret-key-123"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sig := r.Header.Get("X-Guardmesh-Signature")
		receivedSig <- sig

		body, _ := io.ReadAll(r.Body)

		// Verify HMAC
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if sig != expectedSig {
			t.Errorf("signature mismatch: got %s, expected %s", sig, expectedSig)
		}

		var payload WebhookPayload
		_ = json.Unmarshal(body, &payload)
		receivedPayload <- &payload
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	cfg := config.AlertsConfig{
		Enabled: true,
		Webhooks: []config.WebhookTarget{
			{
				Name:   "SlackSIEM",
				URL:    mockServer.URL,
				Secret: secret,
				Events: []string{"blocked", "prohibited"},
			},
		},
	}

	dispatcher := NewDispatcher(cfg)
	defer dispatcher.Close()

	// Dispatch blocked event
	event := &store.AuditEvent{
		ID:               "evt-alert-1",
		Timestamp:        time.Now(),
		AppID:            "crm-portal",
		RiskTier:         "PROHIBITED",
		RiskScore:        1.0,
		RiskReasons:      []string{"Article 5(1)(f): Workplace emotion recognition"},
		IsBlocked:        true,
		BlockedReason:    "Prohibited AI practice",
		ClientIP:         "10.0.0.1",
		PromptSHA256:     "sha-hash-123",
		SampleSnippet:    "Monitor employee video",
		PIIDetectedCount: 0,
	}

	dispatcher.Dispatch(event)

	select {
	case payload := <-receivedPayload:
		if payload.AppID != "crm-portal" {
			t.Errorf("expected crm-portal, got %s", payload.AppID)
		}
		if !payload.IsBlocked {
			t.Errorf("expected IsBlocked true")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for webhook dispatch")
	}
}
