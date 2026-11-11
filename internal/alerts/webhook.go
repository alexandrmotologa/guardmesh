package alerts

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/store"
)

type WebhookPayload struct {
	Event         string    `json:"event"`
	Timestamp     time.Time `json:"timestamp"`
	AppID         string    `json:"app_id"`
	RiskTier      string    `json:"risk_tier"`
	RiskScore     float64   `json:"risk_score"`
	Reasons       []string  `json:"reasons"`
	ClientIP      string    `json:"client_ip"`
	IsBlocked     bool      `json:"is_blocked"`
	BlockedReason string    `json:"blocked_reason,omitempty"`
	PromptSHA256  string    `json:"prompt_sha256"`
	Snippet       string    `json:"snippet"`
	Text          string    `json:"text"` // Native Slack / Discord markdown card
}

type Dispatcher struct {
	cfg     config.AlertsConfig
	client  *http.Client
	queue   chan *WebhookPayload
	stopCh  chan struct{}
}

func NewDispatcher(cfg config.AlertsConfig) *Dispatcher {
	d := &Dispatcher{
		cfg: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		queue:  make(chan *WebhookPayload, 500),
		stopCh: make(chan struct{}),
	}

	if cfg.Enabled && len(cfg.Webhooks) > 0 {
		go d.worker()
	}

	return d
}

func (d *Dispatcher) Close() {
	close(d.stopCh)
}

func (d *Dispatcher) Dispatch(event *store.AuditEvent) {
	if !d.cfg.Enabled || len(d.cfg.Webhooks) == 0 {
		return
	}

	eventType := "audit.event"
	if event.IsBlocked {
		eventType = "violation.blocked"
	} else if event.RiskTier == "PROHIBITED" {
		eventType = "violation.prohibited"
	} else if event.RiskTier == "HIGH_RISK" {
		eventType = "compliance.high_risk"
	}

	// Format Slack / Teams compatible card
	var alertTitle string
	if event.IsBlocked {
		alertTitle = fmt.Sprintf("🚨 *[guardmesh Alert]* Prohibited AI Practice Blocked (HTTP 403) for App `%s`", event.AppID)
	} else if event.RiskTier == "HIGH_RISK" {
		alertTitle = fmt.Sprintf("⚠️ *[guardmesh Notice]* High-Risk AI System (Annex III) Invocation for App `%s`", event.AppID)
	} else {
		alertTitle = fmt.Sprintf("ℹ️ *[guardmesh Audit]* AI Invocation Recorded for App `%s`", event.AppID)
	}

	slackText := fmt.Sprintf("%s\n*Risk Tier:* %s (Score: %.2f)\n*Reasons:* %s\n*Latency:* %.1fms\n*Fingerprint:* `%s`",
		alertTitle,
		event.RiskTier,
		event.RiskScore,
		strings.Join(event.RiskReasons, ", "),
		event.LatencyMs,
		event.PromptSHA256,
	)

	payload := &WebhookPayload{
		Event:         eventType,
		Timestamp:     event.Timestamp,
		AppID:         event.AppID,
		RiskTier:      event.RiskTier,
		RiskScore:     event.RiskScore,
		Reasons:       event.RiskReasons,
		ClientIP:      event.ClientIP,
		IsBlocked:     event.IsBlocked,
		BlockedReason: event.BlockedReason,
		PromptSHA256:  event.PromptSHA256,
		Snippet:       event.SampleSnippet,
		Text:          slackText,
	}

	select {
	case d.queue <- payload:
	default:
		// Queue full, drop to prevent memory leak
	}
}

func (d *Dispatcher) worker() {
	for {
		select {
		case payload := <-d.queue:
			d.sendToTargets(payload)
		case <-d.stopCh:
			return
		}
	}
}

func (d *Dispatcher) sendToTargets(payload *WebhookPayload) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, target := range d.cfg.Webhooks {
		if !d.shouldSend(target.Events, payload) {
			continue
		}

		req, err := http.NewRequest("POST", target.URL, bytes.NewReader(data))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "guardmesh-alert-dispatcher/1.0")

		if target.Secret != "" {
			mac := hmac.New(sha256.New, []byte(target.Secret))
			mac.Write(data)
			sig := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-Guardmesh-Signature", sig)
		}

		resp, err := d.client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
		}
	}
}

func (d *Dispatcher) shouldSend(configuredEvents []string, payload *WebhookPayload) bool {
	if len(configuredEvents) == 0 {
		return true
	}

	for _, e := range configuredEvents {
		e = strings.ToLower(e)
		if e == "all" {
			return true
		}
		if e == "blocked" && payload.IsBlocked {
			return true
		}
		if e == "prohibited" && payload.RiskTier == "PROHIBITED" {
			return true
		}
		if e == "high_risk" && payload.RiskTier == "HIGH_RISK" {
			return true
		}
	}
	return false
}
