package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/classifier"
	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/dossier"
	"github.com/alexandrmotologa/guardmesh/internal/pii"
	"github.com/alexandrmotologa/guardmesh/internal/store"
)

type Server struct {
	cfg       *config.Config
	store     *store.Store
	vault     *pii.Vault
	detector  *pii.Detector
	policy    *classifier.PolicyEngine
	openai    *OpenAIAdapter
	anthropic *AnthropicAdapter
	mux       *http.ServeMux
	webFS     fs.FS
}

func NewServer(cfg *config.Config, s *store.Store, webFS fs.FS) (*Server, error) {
	v, err := pii.NewVault(time.Duration(cfg.PII.TokenTTLSeconds) * time.Second)
	if err != nil {
		return nil, fmt.Errorf("init pii vault: %w", err)
	}

	d := pii.NewDetector(cfg.PII.Detectors)
	p := classifier.NewPolicyEngine(cfg)

	srv := &Server{
		cfg:       cfg,
		store:     s,
		vault:     v,
		detector:  d,
		policy:    p,
		openai:    NewOpenAIAdapter(cfg, v, d, p, s),
		anthropic: NewAnthropicAdapter(cfg, v, d, p, s),
		mux:       http.NewServeMux(),
		webFS:     webFS,
	}

	srv.setupRoutes()
	return srv, nil
}

func (s *Server) setupRoutes() {
	// OpenAI & Anthropic API proxy endpoints
	s.mux.HandleFunc("/v1/chat/completions", s.cors(s.openai.HandleChatCompletions))
	s.mux.HandleFunc("/v1/messages", s.cors(s.anthropic.HandleMessages))
	s.mux.HandleFunc("/v1/models", s.cors(s.handleModelsProxy))

	// Compliance & Governance APIs
	s.mux.HandleFunc("/api/stats", s.cors(s.handleGetStats))
	s.mux.HandleFunc("/api/events", s.cors(s.handleGetEvents))
	s.mux.HandleFunc("/api/apps", s.cors(s.handleGetApps))
	s.mux.HandleFunc("/api/classify", s.cors(s.handleClassify))
	s.mux.HandleFunc("/api/scan", s.cors(s.handleScan))
	s.mux.HandleFunc("/api/apps/", s.cors(s.handleAppSubroutes))

	// Embedded Web UI
	if s.webFS != nil {
		fileServer := http.FileServer(http.FS(s.webFS))
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/v1/") {
				http.NotFound(w, r)
				return
			}
			// Serve static asset or fallback to index.html for SPA routing
			f, err := s.webFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
			// Fallback to index.html
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}
}

func (s *Server) cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Guardmesh-App-ID, x-api-key, anthropic-version")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleModelsProxy(w http.ResponseWriter, r *http.Request) {
	upstreamURL := fmt.Sprintf("%s/v1/models", strings.TrimRight(s.cfg.Server.UpstreamOpenAI, "/"))
	req, err := http.NewRequestWithContext(r.Context(), "GET", upstreamURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", r.Header.Get("Authorization"))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.GetGlobalStats()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := store.AuditFilter{
		AppID:    q.Get("app_id"),
		RiskTier: q.Get("risk_tier"),
		Limit:    limit,
		Offset:   offset,
	}

	if blockedStr := q.Get("is_blocked"); blockedStr != "" {
		isBlocked := blockedStr == "true" || blockedStr == "1"
		filter.IsBlocked = &isBlocked
	}

	events, err := s.store.QueryAuditEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(events)
}

func (s *Server) handleGetApps(w http.ResponseWriter, r *http.Request) {
	apps, err := s.store.ListApplications()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apps)
}

func (s *Server) handleClassify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text  string `json:"text"`
		AppID string `json:"app_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	appID := body.AppID
	if appID == "" {
		appID = "auditor"
	}

	riskEval, injection := s.policy.CheckCompliance(appID, body.Text)

	resp := map[string]interface{}{
		"evaluation": riskEval,
		"injection":  injection,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	entities := s.detector.Detect(body.Text)
	masked, count, cats := s.vault.MaskText("scan-preview", body.Text, s.detector, s.cfg.PII.MaskMode)

	resp := map[string]interface{}{
		"entities":   entities,
		"count":      count,
		"categories": cats,
		"masked":     masked,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleAppSubroutes(w http.ResponseWriter, r *http.Request) {
	// Format: /api/apps/{id}/dossier/{format}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	appID := parts[2]
	if len(parts) >= 4 && parts[3] == "dossier" {
		format := "pdf"
		if len(parts) >= 5 {
			format = parts[4]
		}

		app, err := s.store.GetApplication(appID)
		if err != nil || app == nil {
			// If app not registered yet, create default entry
			app = &store.Application{
				ID:               appID,
				Name:             appID,
				HighestRiskTier:  "MINIMAL_RISK",
				ComplianceStatus: "COMPLIANT",
			}
		}

		events, _ := s.store.QueryAuditEvents(store.AuditFilter{AppID: appID, Limit: 10})
		eval, _ := s.policy.CheckCompliance(appID, app.Name)

		meta := dossier.DossierMetadata{
			AppID:              app.ID,
			SystemName:         app.Name,
			Organization:       "Enterprise AI Governance",
			Author:             "Compliance Officer",
			IntendedPurpose:    "Enterprise AI Production Workload",
			ModelUsed:          "gpt-4o / claude-3-5-sonnet",
			GeneratedAt:        time.Now(),
			RiskEvaluation:     eval,
			TotalRequests:      app.TotalRequests,
			TotalPIIMasked:     app.TotalPIIMasked,
			ComplianceStatus:   app.ComplianceStatus,
			SampleAuditRecords: events,
		}

		if format == "md" || format == "markdown" {
			md := dossier.GenerateAnnexIVMarkdown(meta)
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=EU_AI_Act_Annex_IV_%s.md", appID))
			_, _ = w.Write([]byte(md))
			return
		}

		// PDF format:
		pdfBytes, err := dossier.GeneratePDFBytes(meta)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate PDF dossier: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=EU_AI_Act_Annex_IV_%s.pdf", appID))
		_, _ = w.Write(pdfBytes)
		return
	}

	http.NotFound(w, r)
}
