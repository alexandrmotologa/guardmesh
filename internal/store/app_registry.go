package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Application struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
	HighestRiskTier  string    `json:"highest_risk_tier"`
	TotalRequests    int64     `json:"total_requests"`
	TotalPIIMasked   int64     `json:"total_pii_masked"`
	ComplianceStatus string    `json:"compliance_status"` // "COMPLIANT", "NEEDS_REVIEW", "VIOLATION"
}

type DossierRecord struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	GeneratedAt time.Time `json:"generated_at"`
	RiskTier    string    `json:"risk_tier"`
	FilePath    string    `json:"file_path"`
	Summary     string    `json:"summary"`
}

func (s *Store) upsertApplicationStats(appID, riskTier string, piiCount int, eventTime time.Time) error {
	if appID == "" {
		appID = "default-app"
	}

	// Determine new compliance status based on risk
	complianceStatus := "COMPLIANT"
	if riskTier == "PROHIBITED" {
		complianceStatus = "VIOLATION"
	} else if riskTier == "HIGH_RISK" {
		complianceStatus = "NEEDS_REVIEW"
	}

	query := `
	INSERT INTO applications (id, name, first_seen, last_seen, highest_risk_tier, total_requests, total_pii_masked, compliance_status)
	VALUES (?, ?, ?, ?, ?, 1, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		last_seen = excluded.last_seen,
		total_requests = total_requests + 1,
		total_pii_masked = total_pii_masked + excluded.total_pii_masked,
		highest_risk_tier = CASE
			WHEN excluded.highest_risk_tier = 'PROHIBITED' THEN 'PROHIBITED'
			WHEN highest_risk_tier = 'PROHIBITED' THEN 'PROHIBITED'
			WHEN excluded.highest_risk_tier = 'HIGH_RISK' THEN 'HIGH_RISK'
			WHEN highest_risk_tier = 'HIGH_RISK' THEN 'HIGH_RISK'
			WHEN excluded.highest_risk_tier = 'LIMITED_RISK' THEN 'LIMITED_RISK'
			ELSE highest_risk_tier
		END,
		compliance_status = CASE
			WHEN excluded.compliance_status = 'VIOLATION' THEN 'VIOLATION'
			WHEN compliance_status = 'VIOLATION' THEN 'VIOLATION'
			WHEN excluded.compliance_status = 'NEEDS_REVIEW' THEN 'NEEDS_REVIEW'
			ELSE compliance_status
		END;
	`

	ts := eventTime.Format(time.RFC3339)
	_, err := s.db.Exec(query, appID, appID, ts, ts, riskTier, piiCount, complianceStatus)
	return err
}

func (s *Store) ListApplications() ([]Application, error) {
	rows, err := s.db.Query(`
		SELECT id, name, first_seen, last_seen, highest_risk_tier, total_requests, total_pii_masked, compliance_status
		FROM applications
		ORDER BY last_seen DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	var apps []Application
	for rows.Next() {
		var a Application
		var fsStr, lsStr string
		if err := rows.Scan(&a.ID, &a.Name, &fsStr, &lsStr, &a.HighestRiskTier, &a.TotalRequests, &a.TotalPIIMasked, &a.ComplianceStatus); err != nil {
			return nil, err
		}
		if t, err := time.Parse(time.RFC3339Nano, fsStr); err == nil {
			a.FirstSeen = t
		} else if t, err := time.Parse(time.RFC3339, fsStr); err == nil {
			a.FirstSeen = t
		} else if t, err := time.Parse("2006-01-02T15:04:05.999999", fsStr); err == nil {
			a.FirstSeen = t
		} else if t, err := time.Parse("2006-01-02T15:04:05", fsStr); err == nil {
			a.FirstSeen = t
		}

		if t, err := time.Parse(time.RFC3339Nano, lsStr); err == nil {
			a.LastSeen = t
		} else if t, err := time.Parse(time.RFC3339, lsStr); err == nil {
			a.LastSeen = t
		} else if t, err := time.Parse("2006-01-02T15:04:05.999999", lsStr); err == nil {
			a.LastSeen = t
		} else if t, err := time.Parse("2006-01-02T15:04:05", lsStr); err == nil {
			a.LastSeen = t
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func (s *Store) GetApplication(appID string) (*Application, error) {
	row := s.db.QueryRow(`
		SELECT id, name, first_seen, last_seen, highest_risk_tier, total_requests, total_pii_masked, compliance_status
		FROM applications
		WHERE id = ?
	`, appID)

	var a Application
	var fsStr, lsStr string
	err := row.Scan(&a.ID, &a.Name, &fsStr, &lsStr, &a.HighestRiskTier, &a.TotalRequests, &a.TotalPIIMasked, &a.ComplianceStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	a.FirstSeen, _ = time.Parse(time.RFC3339, fsStr)
	a.LastSeen, _ = time.Parse(time.RFC3339, lsStr)
	return &a, nil
}

func (s *Store) RecordDossier(record *DossierRecord) error {
	query := `
	INSERT INTO dossier_records (id, app_id, generated_at, risk_tier, file_path, summary)
	VALUES (?, ?, ?, ?, ?, ?);
	`
	_, err := s.db.Exec(query, record.ID, record.AppID, record.GeneratedAt.Format(time.RFC3339), record.RiskTier, record.FilePath, record.Summary)
	return err
}

func (s *Store) ListDossiers(appID string) ([]DossierRecord, error) {
	query := `SELECT id, app_id, generated_at, risk_tier, file_path, summary FROM dossier_records`
	var args []interface{}
	if appID != "" {
		query += " WHERE app_id = ?"
		args = append(args, appID)
	}
	query += " ORDER BY generated_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dossiers []DossierRecord
	for rows.Next() {
		var d DossierRecord
		var tsStr string
		if err := rows.Scan(&d.ID, &d.AppID, &tsStr, &d.RiskTier, &d.FilePath, &d.Summary); err != nil {
			return nil, err
		}
		d.GeneratedAt, _ = time.Parse(time.RFC3339, tsStr)
		dossiers = append(dossiers, d)
	}
	return dossiers, nil
}
