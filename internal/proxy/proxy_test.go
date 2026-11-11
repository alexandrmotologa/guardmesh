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
