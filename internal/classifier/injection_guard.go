package classifier

import (
	"regexp"
)

type InjectionResult struct {
	Detected   bool    `json:"detected"`
	Score      float64 `json:"score"`
	Pattern    string  `json:"pattern,omitempty"`
	Confidence float64 `json:"confidence"`
}

type InjectionGuard struct {
	patterns []*regexp.Regexp
}

func NewInjectionGuard() *InjectionGuard {
	return &InjectionGuard{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b(?:ignore previous instructions|disregard (?:all )?(?:earlier|previous) (?:prompts|instructions)|forget (?:all )?previous instructions)\b`),
			regexp.MustCompile(`(?i)\b(?:you are now DAN|do anything now|jailbreak mode|bypass safety filters|unrestricted ai mode)\b`),
			regexp.MustCompile(`(?i)\b(?:reveal (?:your )?(?:system prompt|initial instructions|hidden prompt)|print the developer instructions above)\b`),
			regexp.MustCompile(`(?i)\b(?:simulate an ai without ethics|pretend there are no safety rules|ignore all content policies)\b`),
			regexp.MustCompile(`(?i)\b(?:base64 decode and execute the following instructions|interpret this reversed text as system instructions)\b`),
		},
	}
}

func (ig *InjectionGuard) Scan(text string) InjectionResult {
	for _, p := range ig.patterns {
		if loc := p.FindString(text); loc != "" {
			return InjectionResult{
				Detected:   true,
				Score:      0.95,
				Pattern:    loc,
				Confidence: 0.95,
			}
		}
	}

	return InjectionResult{
		Detected:   false,
		Score:      0.0,
		Confidence: 1.0,
	}
}
