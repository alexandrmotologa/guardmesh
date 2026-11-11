package store

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type AuditEvent struct {
	ID               string    `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	AppID            string    `json:"app_id"`
	ClientIP         string    `json:"client_ip"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	StatusCode       int       `json:"status_code"`
	LatencyMs        float64   `json:"latency_ms"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	RiskTier         string    `json:"risk_tier"`
	RiskScore        float64   `json:"risk_score"`
	RiskReasons      []string  `json:"risk_reasons"`
	IsBlocked        bool      `json:"is_blocked"`
	BlockedReason    string    `json:"blocked_reason,omitempty"`
	PIIDetectedCount int       `json:"pii_detected_count"`
	PIICategories    []string  `json:"pii_categories"`
	PromptSHA256     string    `json:"prompt_sha256"`
	SampleSnippet    string    `json:"sample_snippet"`
}

type AuditFilter struct {
	AppID     string
	RiskTier  string
	IsBlocked *bool
	Limit     int
	Offset    int
}

type GlobalStats struct {
	TotalRequests    int64            `json:"total_requests"`
	TotalBlocked     int64            `json:"total_blocked"`
	TotalPIIMasked   int64            `json:"total_pii_masked"`
	AvgLatencyMs     float64          `json:"avg_latency_ms"`
	RiskDistribution map[string]int64 `json:"risk_distribution"`
	PIIDistribution  map[string]int64 `json:"pii_distribution"`
}

func (s *Store) RecordAuditEvent(e *AuditEvent) error {
	reasonsJSON, err := json.Marshal(e.RiskReasons)
	if err != nil {
		reasonsJSON = []byte("[]")
	}
	catsJSON, err := json.Marshal(e.PIICategories)
	if err != nil {
		catsJSON = []byte("[]")
	}

	query := `
	INSERT INTO audit_events (
		id, timestamp, app_id, client_ip, provider, model,
		status_code, latency_ms, prompt_tokens, completion_tokens,
		risk_tier, risk_score, risk_reasons, is_blocked, blocked_reason,
		pii_detected_count, pii_categories, prompt_sha256, sample_snippet
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	isBlockedInt := 0
	if e.IsBlocked {
		isBlockedInt = 1
	}

	_, err = s.db.Exec(
		query,
		e.ID, e.Timestamp.Format(time.RFC3339Nano), e.AppID, e.ClientIP, e.Provider, e.Model,
		e.StatusCode, e.LatencyMs, e.PromptTokens, e.CompletionTokens,
		e.RiskTier, e.RiskScore, string(reasonsJSON), isBlockedInt, e.BlockedReason,
		e.PIIDetectedCount, string(catsJSON), e.PromptSHA256, e.SampleSnippet,
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}

	// Update Application registry summary
	return s.upsertApplicationStats(e.AppID, e.RiskTier, e.PIIDetectedCount, e.Timestamp)
}

func (s *Store) QueryAuditEvents(filter AuditFilter) ([]AuditEvent, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	query := `
	SELECT id, timestamp, app_id, client_ip, provider, model,
	       status_code, latency_ms, prompt_tokens, completion_tokens,
	       risk_tier, risk_score, risk_reasons, is_blocked, blocked_reason,
	       pii_detected_count, pii_categories, prompt_sha256, sample_snippet
	FROM audit_events
	WHERE 1=1
	`
	var args []interface{}

	if filter.AppID != "" {
		query += " AND app_id = ?"
		args = append(args, filter.AppID)
	}
	if filter.RiskTier != "" {
		query += " AND risk_tier = ?"
		args = append(args, filter.RiskTier)
	}
	if filter.IsBlocked != nil {
		query += " AND is_blocked = ?"
		if *filter.IsBlocked {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		var tsStr, reasonsStr, catsStr string
		var isBlockedInt int
		var blockedReason sql.NullString

		err := rows.Scan(
			&e.ID, &tsStr, &e.AppID, &e.ClientIP, &e.Provider, &e.Model,
			&e.StatusCode, &e.LatencyMs, &e.PromptTokens, &e.CompletionTokens,
			&e.RiskTier, &e.RiskScore, &reasonsStr, &isBlockedInt, &blockedReason,
			&e.PIIDetectedCount, &catsStr, &e.PromptSHA256, &e.SampleSnippet,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}

		if blockedReason.Valid {
			e.BlockedReason = blockedReason.String
		}

		if parsedTime, err := time.Parse(time.RFC3339Nano, tsStr); err == nil {
			e.Timestamp = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, tsStr); err == nil {
			e.Timestamp = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02T15:04:05.999999", tsStr); err == nil {
			e.Timestamp = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02T15:04:05", tsStr); err == nil {
			e.Timestamp = parsedTime
		}

		e.IsBlocked = isBlockedInt == 1
		_ = json.Unmarshal([]byte(reasonsStr), &e.RiskReasons)
		_ = json.Unmarshal([]byte(catsStr), &e.PIICategories)

		events = append(events, e)
	}

	return events, nil
}

func (s *Store) GetGlobalStats() (*GlobalStats, error) {
	stats := &GlobalStats{
		RiskDistribution: make(map[string]int64),
		PIIDistribution:  make(map[string]int64),
	}

	// Total and averages
	var avgLatency sql.NullFloat64
	row := s.db.QueryRow(`
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN is_blocked = 1 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(pii_detected_count), 0),
		       AVG(latency_ms)
		FROM audit_events
	`)
	if err := row.Scan(&stats.TotalRequests, &stats.TotalBlocked, &stats.TotalPIIMasked, &avgLatency); err != nil {
		return nil, fmt.Errorf("scan global stats: %w", err)
	}
	if avgLatency.Valid {
		stats.AvgLatencyMs = avgLatency.Float64
	}

	// Risk distribution
	rows, err := s.db.Query(`SELECT risk_tier, COUNT(*) FROM audit_events GROUP BY risk_tier`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tier string
			var count int64
			if err := rows.Scan(&tier, &count); err == nil {
				stats.RiskDistribution[tier] = count
			}
		}
	}

	return stats, nil
}

func (s *Store) ExportAuditEventsCSV(filter AuditFilter, w io.Writer) error {
	filter.Limit = 100000 // allow exporting all matching records
	events, err := s.QueryAuditEvents(filter)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	header := []string{
		"id", "timestamp", "app_id", "client_ip", "provider", "model",
		"status_code", "latency_ms", "prompt_tokens", "completion_tokens",
		"risk_tier", "risk_score", "is_blocked", "blocked_reason",
		"pii_detected_count", "pii_categories", "prompt_sha256", "sample_snippet",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, e := range events {
		record := []string{
			e.ID,
			e.Timestamp.Format(time.RFC3339),
			e.AppID,
			e.ClientIP,
			e.Provider,
			e.Model,
			strconv.Itoa(e.StatusCode),
			fmt.Sprintf("%.2f", e.LatencyMs),
			strconv.Itoa(e.PromptTokens),
			strconv.Itoa(e.CompletionTokens),
			e.RiskTier,
			fmt.Sprintf("%.2f", e.RiskScore),
			strconv.FormatBool(e.IsBlocked),
			e.BlockedReason,
			strconv.Itoa(e.PIIDetectedCount),
			strings.Join(e.PIICategories, ";"),
			e.PromptSHA256,
			e.SampleSnippet,
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ExportAuditEventsJSONL(filter AuditFilter, w io.Writer) error {
	filter.Limit = 100000
	events, err := s.QueryAuditEvents(filter)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	for _, e := range events {
		if err := encoder.Encode(e); err != nil {
			return err
		}
	}
	return nil
}
