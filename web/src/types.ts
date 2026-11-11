export type RiskTier = 'PROHIBITED' | 'HIGH_RISK' | 'LIMITED_RISK' | 'MINIMAL_RISK';

export interface AuditEvent {
  id: string;
  timestamp: string;
  app_id: string;
  client_ip: string;
  provider: string;
  model: string;
  status_code: number;
  latency_ms: number;
  prompt_tokens: number;
  completion_tokens: number;
  risk_tier: RiskTier;
  risk_score: number;
  risk_reasons: string[];
  is_blocked: boolean;
  blocked_reason?: string;
  pii_detected_count: number;
  pii_categories: string[];
  prompt_sha256: string;
  sample_snippet: string;
}

export interface Application {
  id: string;
  name: string;
  first_seen: string;
  last_seen: string;
  highest_risk_tier: RiskTier;
  total_requests: number;
  total_pii_masked: number;
  compliance_status: 'COMPLIANT' | 'NEEDS_REVIEW' | 'VIOLATION';
}

export interface GlobalStats {
  total_requests: number;
  total_blocked: number;
  total_pii_masked: number;
  avg_latency_ms: number;
  risk_distribution: Record<string, number>;
  pii_distribution: Record<string, number>;
}

export interface RiskEvaluation {
  tier: RiskTier;
  score: number;
  reasons: string[];
  applicable_articles: string[];
  required_controls: string[];
  is_blocked: boolean;
  blocked_reason?: string;
}

export interface InjectionResult {
  detected: boolean;
  score: number;
  pattern?: string;
  confidence: number;
}

export interface ClassifyResponse {
  evaluation: RiskEvaluation;
  injection: InjectionResult;
}

export interface DetectedEntity {
  Category: string;
  Value: string;
  Start: number;
  End: number;
  Confidence: number;
}

export interface ScanResponse {
  entities: DetectedEntity[];
  count: number;
  categories: string[];
  masked: string;
}
