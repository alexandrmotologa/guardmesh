package proxy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/alerts"
	"github.com/alexandrmotologa/guardmesh/internal/classifier"
	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/pii"
	"github.com/alexandrmotologa/guardmesh/internal/store"
	"github.com/google/uuid"
)

type AnthropicAdapter struct {
	cfg      *config.Config
	vault    *pii.Vault
	detector *pii.Detector
	policy   *classifier.PolicyEngine
	store    *store.Store
	alerts   *alerts.Dispatcher
	client   *http.Client
}

func NewAnthropicAdapter(cfg *config.Config, v *pii.Vault, d *pii.Detector, p *classifier.PolicyEngine, s *store.Store, al *alerts.Dispatcher) *AnthropicAdapter {
	return &AnthropicAdapter{
		cfg:      cfg,
		vault:    v,
		detector: d,
		policy:   p,
		store:    s,
		alerts:   al,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	System    interface{}        `json:"system,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
	MaxTokens int                `json:"max_tokens,omitempty"`
	Tools     []interface{}      `json:"tools,omitempty"`
}

func (aa *AnthropicAdapter) HandleMessages(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := "gm-" + uuid.New().String()

	appID := r.Header.Get("X-Guardmesh-App-ID")
	if appID == "" {
		appID = aa.cfg.Governance.DefaultApplicationID
	}
	if appID == "" {
		appID = "default-app"
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"type":"error","error":{"type":"invalid_request_error","message":"Failed to read request body"}}`, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	var req AnthropicRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, `{"type":"error","error":{"type":"invalid_request_error","message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	// 1. Extract context text
	var contextBuilder strings.Builder
	if sysStr, ok := req.System.(string); ok {
		contextBuilder.WriteString(sysStr)
		contextBuilder.WriteString(" ")
	}
	for _, msg := range req.Messages {
		if text, ok := msg.Content.(string); ok {
			contextBuilder.WriteString(text)
			contextBuilder.WriteString(" ")
		}
	}
	if len(req.Tools) > 0 {
		toolsJSON, _ := json.Marshal(req.Tools)
		contextBuilder.Write(toolsJSON)
	}

	contextText := contextBuilder.String()
	promptHashBytes := sha256.Sum256([]byte(contextText))
	promptSHA := hex.EncodeToString(promptHashBytes[:])

	// 2. Evaluate Compliance & Risk
	riskEval, _ := aa.policy.CheckCompliance(appID, contextText)

	// 3. Block if Article 5 Prohibited
	if riskEval.IsBlocked {
		duration := time.Since(start).Seconds() * 1000
		evt := &store.AuditEvent{
			ID:               requestID,
			Timestamp:        start,
			AppID:            appID,
			ClientIP:         r.RemoteAddr,
			Provider:         "anthropic",
			Model:            req.Model,
			StatusCode:       http.StatusForbidden,
			LatencyMs:        duration,
			RiskTier:         string(riskEval.Tier),
			RiskScore:        riskEval.Score,
			RiskReasons:      riskEval.Reasons,
			IsBlocked:        true,
			BlockedReason:    riskEval.BlockedReason,
			PIIDetectedCount: 0,
			PromptSHA256:     promptSHA,
			SampleSnippet:    truncateSnippet(contextText, 120),
		}
		_ = aa.store.RecordAuditEvent(evt)
		if aa.alerts != nil {
			aa.alerts.Dispatch(evt)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Guardmesh-Risk-Tier", string(riskEval.Tier))
		w.Header().Set("X-Guardmesh-Blocked", "true")
		w.WriteHeader(http.StatusForbidden)

		errResp := map[string]interface{}{
			"type": "error",
			"error": map[string]interface{}{
				"type":                "compliance_violation",
				"message":             riskEval.BlockedReason,
				"risk_tier":           riskEval.Tier,
				"applicable_articles": riskEval.ApplicableArticles,
				"reasons":             riskEval.Reasons,
			},
		}
		_ = json.NewEncoder(w).Encode(errResp)
		return
	}

	// 4. De-identify PII
	totalPIICount := 0
	var allCategories []string
	if aa.cfg.PII.Enabled {
		if sysStr, ok := req.System.(string); ok && sysStr != "" {
			masked, count, cats := aa.vault.MaskText(requestID, sysStr, aa.detector, aa.cfg.PII.MaskMode)
			if count > 0 {
				totalPIICount += count
				allCategories = append(allCategories, cats...)
				req.System = masked
			}
		}
		for i, msg := range req.Messages {
			if text, ok := msg.Content.(string); ok && text != "" {
				masked, count, cats := aa.vault.MaskText(requestID, text, aa.detector, aa.cfg.PII.MaskMode)
				if count > 0 {
					totalPIICount += count
					allCategories = append(allCategories, cats...)
					req.Messages[i].Content = masked
				}
			}
		}
	}

	// 5. Forward to Upstream
	sanitizedBody, err := json.Marshal(req)
	if err != nil {
		http.Error(w, `{"type":"error","error":{"type":"api_error","message":"Failed to serialize request"}}`, http.StatusInternalServerError)
		return
	}

	upstreamURL := fmt.Sprintf("%s/v1/messages", strings.TrimRight(aa.cfg.Server.UpstreamAnthropic, "/"))
	upstreamReq, err := http.NewRequestWithContext(r.Context(), "POST", upstreamURL, bytes.NewReader(sanitizedBody))
	if err != nil {
		http.Error(w, `{"type":"error","error":{"type":"api_error","message":"Failed to create upstream request"}}`, http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		if strings.EqualFold(key, "Host") || strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, v := range values {
			upstreamReq.Header.Add(key, v)
		}
	}
	upstreamReq.Header.Set("Content-Type", "application/json")

	resp, err := aa.client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"type":"error","error":{"type":"api_error","message":"Upstream error: %s"}}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 6. Handle Stream vs Non-Stream
	isLimitedRisk := riskEval.Tier == classifier.TierLimitedRisk
	article50Active := isLimitedRisk && aa.cfg.Governance.Article50Transparency
	if article50Active && aa.cfg.Governance.Article50Header {
		w.Header().Set("X-AI-Generated", "true")
		w.Header().Set("X-EU-AI-Act-Article", "Article 50(1)")
	}

	if req.Stream {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported by server", http.StatusInternalServerError)
			return
		}

		for k, v := range resp.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}
		w.Header().Set("X-Guardmesh-Risk-Tier", string(riskEval.Tier))
		w.Header().Set("X-Guardmesh-PII-Masked", fmt.Sprintf("%d", totalPIICount))
		w.WriteHeader(resp.StatusCode)

		suffix := ""
		if article50Active && aa.cfg.Governance.Article50DisclaimerText != "" {
			suffix = aa.cfg.Governance.Article50DisclaimerText
		}

		rehydrator := pii.NewStreamRehydrator(requestID, aa.vault)
		tokensCount, _ := InterceptAnthropicSSEStream(resp.Body, w, flusher, rehydrator, suffix)

		duration := time.Since(start).Seconds() * 1000
		evt := &store.AuditEvent{
			ID:               requestID,
			Timestamp:        start,
			AppID:            appID,
			ClientIP:         r.RemoteAddr,
			Provider:         "anthropic",
			Model:            req.Model,
			StatusCode:       resp.StatusCode,
			LatencyMs:        duration,
			PromptTokens:     len(contextText) / 4,
			CompletionTokens: tokensCount,
			RiskTier:         string(riskEval.Tier),
			RiskScore:        riskEval.Score,
			RiskReasons:      riskEval.Reasons,
			IsBlocked:        false,
			PIIDetectedCount: totalPIICount,
			PIICategories:    allCategories,
			PromptSHA256:     promptSHA,
			SampleSnippet:    truncateSnippet(contextText, 120),
		}
		_ = aa.store.RecordAuditEvent(evt)
		if aa.alerts != nil {
			aa.alerts.Dispatch(evt)
		}
		return
	}

	// Non-Streaming Response
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, `{"type":"error","error":{"type":"api_error","message":"Failed to read upstream response"}}`, http.StatusBadGateway)
		return
	}

	var respJSON map[string]interface{}
	completionTokens := 0
	if err := json.Unmarshal(respBytes, &respJSON); err == nil {
		contentBlocks, ok := respJSON["content"].([]interface{})
		if ok {
			for _, block := range contentBlocks {
				if blockMap, ok := block.(map[string]interface{}); ok {
					if blockMap["type"] == "text" {
						if textVal, ok := blockMap["text"].(string); ok {
							rehydrated := aa.vault.RehydrateText(requestID, textVal)
							if article50Active && aa.cfg.Governance.Article50DisclaimerText != "" {
								rehydrated += aa.cfg.Governance.Article50DisclaimerText
							}
							blockMap["text"] = rehydrated
						}
					}
				}
			}
			respBytes, _ = json.Marshal(respJSON)
		}
		if usage, ok := respJSON["usage"].(map[string]interface{}); ok {
			if ct, ok := usage["output_tokens"].(float64); ok {
				completionTokens = int(ct)
			}
		}
	}

	duration := time.Since(start).Seconds() * 1000
	evt := &store.AuditEvent{
		ID:               requestID,
		Timestamp:        start,
		AppID:            appID,
		ClientIP:         r.RemoteAddr,
		Provider:         "anthropic",
		Model:            req.Model,
		StatusCode:       resp.StatusCode,
		LatencyMs:        duration,
		PromptTokens:     len(contextText) / 4,
		CompletionTokens: completionTokens,
		RiskTier:         string(riskEval.Tier),
		RiskScore:        riskEval.Score,
		RiskReasons:      riskEval.Reasons,
		IsBlocked:        false,
		PIIDetectedCount: totalPIICount,
		PIICategories:    allCategories,
		PromptSHA256:     promptSHA,
		SampleSnippet:    truncateSnippet(contextText, 120),
	}
	_ = aa.store.RecordAuditEvent(evt)
	if aa.alerts != nil {
		aa.alerts.Dispatch(evt)
	}

	for k, v := range resp.Header {
		if strings.EqualFold(k, "Content-Length") {
			continue
		}
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.Header().Set("X-Guardmesh-Risk-Tier", string(riskEval.Tier))
	w.Header().Set("X-Guardmesh-PII-Masked", fmt.Sprintf("%d", totalPIICount))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(respBytes)))
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBytes)
}
