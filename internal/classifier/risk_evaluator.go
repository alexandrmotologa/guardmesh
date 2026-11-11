package classifier

import (
	"regexp"
	"strings"

	"github.com/alexandrmotologa/guardmesh/internal/config"
)

type RiskTier string

const (
	TierProhibited  RiskTier = "PROHIBITED"   // Article 5 EU AI Act (Hard Block)
	TierHighRisk    RiskTier = "HIGH_RISK"    // Annex III EU AI Act (Mandatory Dossier & Human-in-the-loop)
	TierLimitedRisk RiskTier = "LIMITED_RISK" // Article 50 EU AI Act (Transparency Obligations)
	TierMinimalRisk RiskTier = "MINIMAL_RISK" // Minimal / No specific obligations
)

type RiskEvaluation struct {
	Tier               RiskTier `json:"tier"`
	Score              float64  `json:"score"`
	Reasons            []string `json:"reasons"`
	ApplicableArticles []string `json:"applicable_articles"`
	RequiredControls   []string `json:"required_controls"`
	IsBlocked          bool     `json:"is_blocked"`
	BlockedReason      string   `json:"blocked_reason,omitempty"`
}

type RulePattern struct {
	Tier      RiskTier
	Article   string
	Reason    string
	Control   string
	Regex     *regexp.Regexp
	BaseScore float64
}

type RiskEvaluator struct {
	cfg   config.GovernanceConfig
	rules []RulePattern
}

func NewRiskEvaluator(cfg config.GovernanceConfig) *RiskEvaluator {
	re := &RiskEvaluator{
		cfg: cfg,
	}
	re.initRules()
	return re
}

func (re *RiskEvaluator) initRules() {
	re.rules = []RulePattern{
		// === ARTICLE 5: PROHIBITED AI PRACTICES ===
		{
			Tier:      TierProhibited,
			Article:   "Article 5(1)(a)",
			Reason:    "Subliminal or manipulative techniques distorting human behavior and causing harm",
			Control:   "Immediate deployment ban",
			Regex:     regexp.MustCompile(`(?i)\b(?:subliminal manipulation|manipulate.*subconsciously|circumvent.*free will|covertly influence.*behavior)\b`),
			BaseScore: 1.0,
		},
		{
			Tier:      TierProhibited,
			Article:   "Article 5(1)(c)",
			Reason:    "Social scoring or citizen trustworthiness evaluation by public/private authorities",
			Control:   "Immediate deployment ban",
			Regex:     regexp.MustCompile(`(?i)\b(?:social scoring|social credit score|citizen rating|rate citizen trustworthiness|social hierarchy score)\b`),
			BaseScore: 1.0,
		},
		{
			Tier:      TierProhibited,
			Article:   "Article 5(1)(d)",
			Reason:    "Individual predictive policing based solely on profiling or personality traits",
			Control:   "Immediate deployment ban",
			Regex:     regexp.MustCompile(`(?i)\b(?:predict criminal behavior|predictive policing.*profile|assess likelihood.*commit crime based on traits)\b`),
			BaseScore: 1.0,
		},
		{
			Tier:      TierProhibited,
			Article:   "Article 5(1)(e)",
			Reason:    "Untargeted scraping of facial images from the internet or CCTV footage",
			Control:   "Immediate deployment ban",
			Regex:     regexp.MustCompile(`(?i)\b(?:scrape facial images|cctv facial scrape|harvest face database|untargeted face collection)\b`),
			BaseScore: 1.0,
		},
		{
			Tier:      TierProhibited,
			Article:   "Article 5(1)(f)",
			Reason:    "Emotion recognition in workplace or educational institutions",
			Control:   "Immediate deployment ban",
			Regex:     regexp.MustCompile(`(?i)\b(?:workplace emotion recognition|employee mood detection|monitor worker sentiment.*camera|classroom emotion tracking|student emotion inference)\b`),
			BaseScore: 1.0,
		},
		{
			Tier:      TierProhibited,
			Article:   "Article 5(1)(g)",
			Reason:    "Biometric categorization deducing protected traits (race, political, religious, sexual)",
			Control:   "Immediate deployment ban",
			Regex:     regexp.MustCompile(`(?i)\b(?:deduce sexual orientation|infer religious beliefs|classify political affiliation.*face|biometric categorization.*race|infer ethnic origin.*biometrics)\b`),
			BaseScore: 1.0,
		},

		// === ANNEX III: HIGH-RISK AI SYSTEMS ===
		// Annex III 4: Employment, Workers Management & Self-Employment
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(4)(a)",
			Reason:    "Recruitment, CV screening, job application filtering or candidate evaluation",
			Control:   "Article 14 Human Oversight, Article 11 Technical Documentation",
			Regex:     regexp.MustCompile(`(?i)\b(?:screen resumes|filter candidates|recruitment screening|shortlist applicants|score job applications|cv ranking|evaluate job interviews)\b`),
			BaseScore: 0.85,
		},
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(4)(b)",
			Reason:    "Workplace allocation, performance evaluation, promotion or termination decisions",
			Control:   "Article 14 Human Oversight, Article 9 Risk Management",
			Regex:     regexp.MustCompile(`(?i)\b(?:workplace performance evaluation|decide termination|promotion ranking|employee productivity surveillance|allocate tasks based on worker profiling)\b`),
			BaseScore: 0.85,
		},
		// Annex III 5: Essential Private and Public Services
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(5)(a)",
			Reason:    "Eligibility assessment for public welfare benefits or healthcare services",
			Control:   "Article 10 Data Governance, Article 14 Human Oversight",
			Regex:     regexp.MustCompile(`(?i)\b(?:welfare benefit eligibility|public assistance allocation|healthcare entitlement assessment|social security eligibility)\b`),
			BaseScore: 0.80,
		},
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(5)(b)",
			Reason:    "Creditworthiness assessment or credit scoring of natural persons",
			Control:   "Article 10 Data Quality, Article 13 Transparency, Article 14 Human Oversight",
			Regex:     regexp.MustCompile(`(?i)\b(?:credit score evaluation|assess creditworthiness|loan approval decision|credit risk score|underwriting decision.*credit)\b`),
			BaseScore: 0.85,
		},
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(5)(c)",
			Reason:    "Risk assessment and pricing for life and health insurance",
			Control:   "Article 10 Data Governance, Bias Auditing",
			Regex:     regexp.MustCompile(`(?i)\b(?:health insurance underwriting|life insurance pricing|insurance risk assessment.*individual|premium pricing based on health data)\b`),
			BaseScore: 0.80,
		},
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(5)(d)",
			Reason:    "Emergency services dispatch or emergency call prioritization",
			Control:   "Article 15 Accuracy & Cybersecurity, High Availability",
			Regex:     regexp.MustCompile(`(?i)\b(?:emergency dispatch triage|prioritize 911 calls|emergency ambulance dispatch|fire brigade dispatch)\b`),
			BaseScore: 0.85,
		},
		// Annex III 3: Education and Vocational Training
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(3)(a)",
			Reason:    "Admission or assignment of students to educational institutions",
			Control:   "Article 14 Human Oversight, Fairness Verification",
			Regex:     regexp.MustCompile(`(?i)\b(?:student admission screening|university admission evaluation|school assignment algorithm)\b`),
			BaseScore: 0.75,
		},
		{
			Tier:      TierHighRisk,
			Article:   "Annex III(3)(b)",
			Reason:    "Evaluating learning outcomes or grading student examinations",
			Control:   "Article 13 Transparency, Human Review of Grades",
			Regex:     regexp.MustCompile(`(?i)\b(?:automated exam grading|grade student essays|evaluate learning outcomes.*proctoring|automated proctoring)\b`),
			BaseScore: 0.75,
		},

		// === ARTICLE 50: LIMITED RISK AI SYSTEMS ===
		{
			Tier:      TierLimitedRisk,
			Article:   "Article 50(1)",
			Reason:    "Direct interaction with natural persons via conversational chatbot",
			Control:   "Mandatory transparency notice (inform user they are interacting with AI)",
			Regex:     regexp.MustCompile(`(?i)\b(?:customer support chatbot|conversational assistant|virtual customer agent|interactive helpdesk ai)\b`),
			BaseScore: 0.40,
		},
		{
			Tier:      TierLimitedRisk,
			Article:   "Article 50(2)",
			Reason:    "Generation or manipulation of synthetic media or deepfakes",
			Control:   "Machine-readable watermarking and disclosure of AI-generated content",
			Regex:     regexp.MustCompile(`(?i)\b(?:generate deepfake|clone voice|synthetic voice generation|generate photorealistic human face|voice synthesis)\b`),
			BaseScore: 0.50,
		},
	}
}

// Evaluate analyzes input text (system prompt, user messages, tool schemas) and produces an EU AI Act classification
func (re *RiskEvaluator) Evaluate(contextText string) RiskEvaluation {
	var matchedRules []RulePattern
	var maxScore float64

	for _, rule := range re.rules {
		if rule.Regex.MatchString(contextText) {
			matchedRules = append(matchedRules, rule)
			if rule.BaseScore > maxScore {
				maxScore = rule.BaseScore
			}
		}
	}

	// Determine overall tier
	tier := TierMinimalRisk
	if len(matchedRules) > 0 {
		// Priority: Prohibited > HighRisk > LimitedRisk > Minimal
		for _, m := range matchedRules {
			if m.Tier == TierProhibited {
				tier = TierProhibited
				break
			}
			if m.Tier == TierHighRisk && tier != TierProhibited {
				tier = TierHighRisk
			}
			if m.Tier == TierLimitedRisk && tier == TierMinimalRisk {
				tier = TierLimitedRisk
			}
		}
	}

	reasonsSet := make(map[string]struct{})
	articlesSet := make(map[string]struct{})
	controlsSet := make(map[string]struct{})

	for _, m := range matchedRules {
		reasonsSet[m.Reason] = struct{}{}
		articlesSet[m.Article] = struct{}{}
		controlsSet[m.Control] = struct{}{}
	}

	var reasons, articles, controls []string
	for r := range reasonsSet {
		reasons = append(reasons, r)
	}
	for a := range articlesSet {
		articles = append(articles, a)
	}
	for c := range controlsSet {
		controls = append(controls, c)
	}

	// Default controls for minimal risk
	if tier == TierMinimalRisk {
		reasons = append(reasons, "Standard generative AI operation; no high-risk Annex III or prohibited triggers detected")
		articles = append(articles, "Article 4 (AI Literacy)")
		controls = append(controls, "Basic monitoring and operational telemetry")
		maxScore = 0.10
	}

	eval := RiskEvaluation{
		Tier:               tier,
		Score:              maxScore,
		Reasons:            reasons,
		ApplicableArticles: articles,
		RequiredControls:   controls,
	}

	if tier == TierProhibited && re.cfg.EnforceArticle5Block {
		eval.IsBlocked = true
		eval.BlockedReason = "Blocked under EU AI Act Article 5: Prohibited AI Practice (" + strings.Join(articles, ", ") + ")"
	}

	return eval
}
