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

	"github.com/alexandrmotologa/guardmesh/internal/classifier"
	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/pii"
	"github.com/alexandrmotologa/guardmesh/internal/store"
	"github.com/google/uuid"
)

type OpenAIAdapter struct {
	cfg      *config.Config
	vault    *pii.Vault
	detector *pii.Detector
	policy   *classifier.PolicyEngine
	store    *store.Store
	client   *http.Client
}

func NewOpenAIAdapter(cfg *config.Config, v *pii.Vault, d *pii.Detector, p *classifier.PolicyEngine, s *store.Store) *OpenAIAdapter {
	return &OpenAIAdapter{
		cfg:      cfg,
		vault:    v,
		detector: d,
		policy:   p,
		store:    s,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type OpenAIChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
	Name    string      `json:"name,omitempty"`
}

type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	Stream      bool                `json:"stream,omitempty"`
	Temperature *float64            `json:"temperature,omitempty"`
	Tools       []interface{}       `json:"tools,omitempty"`
}

func (oa *OpenAIAdapter) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := "gm-" + uuid.New().String()

	appID := r.Header.Get("X-Guardmesh-App-ID")
	if appID == "" {
		appID = oa.cfg.Governance.DefaultApplicationID
	}
	if appID == "" {
		appID = "default-app"
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to read request body"}}`, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	var chatReq OpenAIChatRequest
	if err := json.Unmarshal(bodyBytes, &chatReq); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON payload"}}`, http.StatusBadRequest)
		return
	}

	// 1. Extract context text for EU AI Act classification & prompt injection scan
	var contextBuilder strings.Builder
	for _, msg := range chatReq.Messages {
		if text, ok := msg.Content.(string); ok {
			contextBuilder.WriteString(text)
			contextBuilder.WriteString(" ")
		}
	}
	if len(chatReq.Tools) > 0 {
		toolsJSON, _ := json.Marshal(chatReq.Tools)
		contextBuilder.Write(toolsJSON)
	}

	contextText := contextBuilder.String()
	promptHashBytes := sha256.Sum256([]byte(contextText))
	promptSHA := hex.EncodeToString(promptHashBytes[:])

	// 2. Evaluate Compliance & Risk
	riskEval, injection := oa.policy.CheckCompliance(appID, contextText)

	// 3. Check if request must be blocked (Article 5 Prohibited or Prompt Injection Block)
	if riskEval.IsBlocked {
		duration := time.Since(start).Seconds() * 1000
		_ = oa.store.RecordAuditEvent(&store.AuditEvent{
			ID:               requestID,
			Timestamp:        start,
			AppID:            appID,
			ClientIP:         r.RemoteAddr,
			Provider:         "openai",
			Model:            chatReq.Model,
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
		})

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Guardmesh-Risk-Tier", string(riskEval.Tier))
		w.Header().Set("X-Guardmesh-Blocked", "true")
		w.WriteHeader(http.StatusForbidden)

		errResp := map[string]interface{}{
			"error": map[string]interface{}{
				"message":             riskEval.BlockedReason,
				"type":                "compliance_violation",
				"code":                "eu_ai_act_prohibited",
				"risk_tier":           riskEval.Tier,
				"applicable_articles": riskEval.ApplicableArticles,
				"reasons":             riskEval.Reasons,
			},
		}
		_ = json.NewEncoder(w).Encode(errResp)
		return
	}

	// 4. De-identify PII in messages
	totalPIICount := 0
	var allCategories []string
	if oa.cfg.PII.Enabled {
		for i, msg := range chatReq.Messages {
			if text, ok := msg.Content.(string); ok && text != "" {
				masked, count, cats := oa.vault.MaskText(requestID, text, oa.detector, oa.cfg.PII.MaskMode)
				if count > 0 {
					totalPIICount += count
					allCategories = append(allCategories, cats...)
					chatReq.Messages[i].Content = masked
				}
			}
		}
	}

	// 5. Serialize sanitized payload for upstream
	sanitizedBody, err := json.Marshal(chatReq)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to serialize sanitized payload"}}`, http.StatusInternalServerError)
		return
	}

	upstreamURL := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(oa.cfg.Server.UpstreamOpenAI, "/"))
	upstreamReq, err := http.NewRequestWithContext(r.Context(), "POST", upstreamURL, bytes.NewReader(sanitizedBody))
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to create upstream request"}}`, http.StatusInternalServerError)
		return
	}

	// Forward client headers (Authorization, Org, etc.)
	for key, values := range r.Header {
		if strings.EqualFold(key, "Host") || strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, v := range values {
			upstreamReq.Header.Add(key, v)
		}
	}
	upstreamReq.Header.Set("Content-Type", "application/json")

	// 6. Execute Upstream Request
	resp, err := oa.client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":{"message":"Upstream error: %s"}}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 7. Handle Streaming vs Non-Streaming Response
	if chatReq.Stream {
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

		rehydrator := pii.NewStreamRehydrator(requestID, oa.vault)
		tokensCount, _ := InterceptOpenAISSEStream(resp.Body, w, flusher, rehydrator)

		duration := time.Since(start).Seconds() * 1000
		_ = oa.store.RecordAuditEvent(&store.AuditEvent{
			ID:               requestID,
			Timestamp:        start,
			AppID:            appID,
			ClientIP:         r.RemoteAddr,
			Provider:         "openai",
			Model:            chatReq.Model,
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
		})
		return
	}

	// Non-Streaming Response:
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to read upstream response"}}`, http.StatusBadGateway)
		return
	}

	// Re-hydrate tokens in completion JSON
	var respJSON map[string]interface{}
	completionTokens := 0
	if err := json.Unmarshal(respBytes, &respJSON); err == nil {
		choices, ok := respJSON["choices"].([]interface{})
		if ok && len(choices) > 0 {
			choice, ok := choices[0].(map[string]interface{})
			if ok {
				message, ok := choice["message"].(map[string]interface{})
				if ok {
					content, ok := message["content"].(string)
					if ok && content != "" {
						rehydrated := oa.vault.RehydrateText(requestID, content)
						message["content"] = rehydrated
						choice["message"] = message
						choices[0] = choice
						respJSON["choices"] = choices
						respBytes, _ = json.Marshal(respJSON)
					}
				}
			}
		}
		if usage, ok := respJSON["usage"].(map[string]interface{}); ok {
			if ct, ok := usage["completion_tokens"].(float64); ok {
				completionTokens = int(ct)
			}
		}
	}

	duration := time.Since(start).Seconds() * 1000
	_ = oa.store.RecordAuditEvent(&store.AuditEvent{
		ID:               requestID,
		Timestamp:        start,
		AppID:            appID,
		ClientIP:         r.RemoteAddr,
		Provider:         "openai",
		Model:            chatReq.Model,
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
	})

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

	_ = injection // used for security monitoring
}

func truncateSnippet(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
