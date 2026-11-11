package classifier

import (
	"sync"

	"github.com/alexandrmotologa/guardmesh/internal/config"
)

type PolicyEngine struct {
	mu           sync.RWMutex
	evaluator    *RiskEvaluator
	guard        *InjectionGuard
	cfg          config.Config
	appOverrides map[string]RiskTier
}

func NewPolicyEngine(cfg *config.Config) *PolicyEngine {
	return &PolicyEngine{
		evaluator:    NewRiskEvaluator(cfg.Governance),
		guard:        NewInjectionGuard(),
		cfg:          *cfg,
		appOverrides: make(map[string]RiskTier),
	}
}

func (pe *PolicyEngine) SetAppOverride(appID string, tier RiskTier) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.appOverrides[appID] = tier
}

func (pe *PolicyEngine) RemoveAppOverride(appID string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	delete(pe.appOverrides, appID)
}

func (pe *PolicyEngine) CheckCompliance(appID, contextText string) (RiskEvaluation, InjectionResult) {
	pe.mu.RLock()
	override, hasOverride := pe.appOverrides[appID]
	pe.mu.RUnlock()

	eval := pe.evaluator.Evaluate(contextText)
	if hasOverride {
		eval.Tier = override
		if override == TierProhibited && pe.cfg.Governance.EnforceArticle5Block {
			eval.IsBlocked = true
			eval.BlockedReason = "Blocked by enterprise policy override: Application marked as Prohibited"
		} else if override != TierProhibited {
			eval.IsBlocked = false
			eval.BlockedReason = ""
		}
	}

	var injection InjectionResult
	if pe.cfg.Security.ScanPromptInjection {
		injection = pe.guard.Scan(contextText)
		if injection.Detected && pe.cfg.Security.BlockPromptInjection {
			eval.IsBlocked = true
			eval.BlockedReason = "Blocked by security policy: Prompt injection pattern detected (" + injection.Pattern + ")"
		}
	}

	return eval, injection
}
