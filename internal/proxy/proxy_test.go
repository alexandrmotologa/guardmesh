package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/store"
)

func TestProxy_OpenAIPrivateForwarding(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_proxy_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var upstreamReceivedBody []byte

	// 1. Mock upstream OpenAI server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamReceivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-test",
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Received information for {{GUARDMESH_ENTITY_EMAIL_1}} and {{GUARDMESH_ENTITY_SSN_1}} successfully.",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     25,
				"completion_tokens": 12,
				"total_tokens":      37,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	// 2. Set up GuardMesh Proxy
	cfg := config.DefaultConfig()
	cfg.Server.UpstreamOpenAI = mockUpstream.URL
	cfg.Server.DBPath = filepath.Join(tempDir, "proxy_test.db")

	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	srv, err := NewServer(cfg, s, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	proxyServer := httptest.NewServer(srv)
	defer proxyServer.Close()

	// 3. Client calls Proxy with real PII
	clientReqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Please verify user with email alice.smith@acme.com and SSN 123-45-6789.",
			},
		},
		"stream": false,
	}
	reqData, _ := json.Marshal(clientReqBody)

	req, err := http.NewRequest("POST", proxyServer.URL+"/v1/chat/completions", bytes.NewReader(reqData))
	if err != nil {
		t.Fatalf("failed to create client req: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Guardmesh-App-ID", "test-crm")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	// 4. Assert: Upstream NEVER received real PII
	upstreamStr := string(upstreamReceivedBody)
	if strings.Contains(upstreamStr, "alice.smith@acme.com") {
		t.Errorf("LEAK: upstream received real email!")
	}
	if strings.Contains(upstreamStr, "123-45-6789") {
		t.Errorf("LEAK: upstream received real SSN!")
	}
	if !strings.Contains(upstreamStr, "{{GUARDMESH_ENTITY_EMAIL_1}}") {
		t.Errorf("expected email replacement token in upstream request, got: %s", upstreamStr)
	}

	// 5. Assert: Client received re-hydrated response with original PII
	clientRespBytes, _ := io.ReadAll(resp.Body)
	clientRespStr := string(clientRespBytes)
	if !strings.Contains(clientRespStr, "alice.smith@acme.com") {
		t.Errorf("client did not receive rehydrated email: %s", clientRespStr)
	}
	if !strings.Contains(clientRespStr, "123-45-6789") {
		t.Errorf("client did not receive rehydrated SSN: %s", clientRespStr)
	}

	// 6. Assert: Audit event recorded in SQLite
	events, err := s.QueryAuditEvents(store.AuditFilter{AppID: "test-crm"})
	if err != nil || len(events) == 0 {
		t.Fatalf("audit event not recorded in database: %v", err)
	}
	if events[0].PIIDetectedCount < 2 {
		t.Errorf("expected at least 2 PII detected, got %d", events[0].PIIDetectedCount)
	}
}

func TestProxy_Article5ProhibitedHardBlock(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_block_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	upstreamHit := false
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer mockUpstream.Close()

	cfg := config.DefaultConfig()
	cfg.Server.UpstreamOpenAI = mockUpstream.URL
	cfg.Server.DBPath = filepath.Join(tempDir, "block_test.db")
	cfg.Governance.EnforceArticle5Block = true

	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	srv, err := NewServer(cfg, s, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	proxyServer := httptest.NewServer(srv)
	defer proxyServer.Close()

	// Prohibited prompt: Workplace emotion recognition
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Deploy cameras in the office for employee mood detection and workplace emotion recognition.",
			},
		},
	}
	reqData, _ := json.Marshal(reqBody)

	resp, err := http.Post(proxyServer.URL+"/v1/chat/completions", "application/json", bytes.NewReader(reqData))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for Article 5 prohibited practice, got %d", resp.StatusCode)
	}

	if upstreamHit {
		t.Errorf("prohibited request must NEVER reach upstream server!")
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var errResp map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &errResp)

	errMap, ok := errResp["error"].(map[string]interface{})
	if !ok || errMap["code"] != "eu_ai_act_prohibited" {
		t.Errorf("expected code eu_ai_act_prohibited, got: %v", errResp)
	}
}

func TestProxy_Article50Transparency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_art50_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id": "chatcmpl-art50",
			"choices": []interface{}{
				map[string]interface{}{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Hello! I can help you with flight bookings.",
					},
					"finish_reason": "stop",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	cfg := config.DefaultConfig()
	cfg.Server.UpstreamOpenAI = mockUpstream.URL
	cfg.Server.DBPath = filepath.Join(tempDir, "art50.db")
	cfg.Governance.Article50Transparency = true
	cfg.Governance.Article50Header = true
	cfg.Governance.Article50DisclaimerText = "\n\n*(Notice: Output generated by an AI system pursuant to Article 50)*"

	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		t.Fatalf("open store failed: %v", err)
	}
	defer s.Close()

	srv, err := NewServer(cfg, s, nil)
	if err != nil {
		t.Fatalf("new server failed: %v", err)
	}

	proxyServer := httptest.NewServer(srv)
	defer proxyServer.Close()

	// Limited Risk prompt (customer support chatbot)
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a customer support chatbot helping travelers.",
			},
			{
				"role":    "user",
				"content": "Can I change my flight ticket?",
			},
		},
	}
	reqData, _ := json.Marshal(reqBody)

	resp, err := http.Post(proxyServer.URL+"/v1/chat/completions", "application/json", bytes.NewReader(reqData))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify Header
	if resp.Header.Get("X-AI-Generated") != "true" {
		t.Errorf("expected X-AI-Generated header")
	}

	// Verify Disclaimer appended
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "pursuant to Article 50") {
		t.Errorf("expected Article 50 disclaimer in response: %s", string(body))
	}
}

func TestProxy_DossierExportSubroutes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_dossier_route_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.DefaultConfig()
	cfg.Server.DBPath = filepath.Join(tempDir, "route_test.db")

	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	srv, err := NewServer(cfg, s, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	proxyServer := httptest.NewServer(srv)
	defer proxyServer.Close()

	// 1. Download Markdown dossier
	respMd, err := http.Get(proxyServer.URL + "/api/apps/crm-agent/dossier/md")
	if err != nil {
		t.Fatalf("md dossier request failed: %v", err)
	}
	defer respMd.Body.Close()
	if respMd.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for md dossier, got %d", respMd.StatusCode)
	}
	mdContent, _ := io.ReadAll(respMd.Body)
	if !strings.Contains(string(mdContent), "EU AI Act Technical Documentation") {
		t.Errorf("expected md dossier content")
	}

	// 2. Download PDF dossier
	respPdf, err := http.Get(proxyServer.URL + "/api/apps/crm-agent/dossier/pdf")
	if err != nil {
		t.Fatalf("pdf dossier request failed: %v", err)
	}
	defer respPdf.Body.Close()
	if respPdf.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for pdf dossier, got %d", respPdf.StatusCode)
	}
	pdfContent, _ := io.ReadAll(respPdf.Body)
	if len(pdfContent) < 1000 || !strings.HasPrefix(string(pdfContent[:5]), "%PDF-") {
		t.Errorf("expected valid PDF output, got %d bytes", len(pdfContent))
	}
}

func TestProxy_ObservabilityAndExportEndpoints(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "guardmesh_obs_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.DefaultConfig()
	cfg.Server.DBPath = filepath.Join(tempDir, "obs_test.db")

	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	// Seed an audit event
	_ = s.RecordAuditEvent(&store.AuditEvent{
		ID:               "evt-obs-1",
		AppID:            "obs-bot",
		RiskTier:         "MINIMAL_RISK",
		StatusCode:       200,
		LatencyMs:        2.5,
		PIIDetectedCount: 1,
		PromptSHA256:     "sha-obs",
		SampleSnippet:    "Snippet",
	})

	srv, err := NewServer(cfg, s, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()

	proxyServer := httptest.NewServer(srv)
	defer proxyServer.Close()

	// 1. /healthz
	respH, err := http.Get(proxyServer.URL + "/healthz")
	if err != nil || respH.StatusCode != http.StatusOK {
		t.Fatalf("healthz failed: %v", err)
	}
	defer respH.Body.Close()

	// 2. /readyz
	respR, err := http.Get(proxyServer.URL + "/readyz")
	if err != nil || respR.StatusCode != http.StatusOK {
		t.Fatalf("readyz failed: %v", err)
	}
	defer respR.Body.Close()

	// 3. /metrics (Prometheus format)
	respM, err := http.Get(proxyServer.URL + "/metrics")
	if err != nil || respM.StatusCode != http.StatusOK {
		t.Fatalf("metrics failed: %v", err)
	}
	defer respM.Body.Close()
	metricsBytes, _ := io.ReadAll(respM.Body)
	metricsStr := string(metricsBytes)
	if !strings.Contains(metricsStr, "guardmesh_requests_total") {
		t.Errorf("metrics missing guardmesh_requests_total: %s", metricsStr)
	}

	// 4. /api/events/export?format=csv
	respCsv, err := http.Get(proxyServer.URL + "/api/events/export?format=csv")
	if err != nil || respCsv.StatusCode != http.StatusOK {
		t.Fatalf("export csv failed: %v", err)
	}
	defer respCsv.Body.Close()
	csvBytes, _ := io.ReadAll(respCsv.Body)
	if !strings.Contains(string(csvBytes), "id,timestamp,app_id") {
		t.Errorf("unexpected csv output: %s", string(csvBytes))
	}

	// 5. /api/events/export?format=jsonl
	respJsonl, err := http.Get(proxyServer.URL + "/api/events/export?format=jsonl")
	if err != nil || respJsonl.StatusCode != http.StatusOK {
		t.Fatalf("export jsonl failed: %v", err)
	}
	defer respJsonl.Body.Close()
	jsonlBytes, _ := io.ReadAll(respJsonl.Body)
	if !strings.Contains(string(jsonlBytes), `"app_id":"obs-bot"`) {
		t.Errorf("unexpected jsonl output: %s", string(jsonlBytes))
	}
}
