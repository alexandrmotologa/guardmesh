package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	PII        PIIConfig        `yaml:"pii"`
	Governance GovernanceConfig `yaml:"governance"`
	Security   SecurityConfig   `yaml:"security"`
}

type ServerConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	UpstreamOpenAI    string `yaml:"upstream_openai"`
	UpstreamAnthropic string `yaml:"upstream_anthropic"`
	APIKey            string `yaml:"api_key"`
	DBPath            string `yaml:"db_path"`
}

type PIIConfig struct {
	Enabled         bool             `yaml:"enabled"`
	MaskMode        string           `yaml:"mask_mode"` // "token", "redact", "hash"
	TokenTTLSeconds int              `yaml:"token_ttl_seconds"`
	Detectors       DetectorsConfig  `yaml:"detectors"`
}

type DetectorsConfig struct {
	CreditCard bool `yaml:"credit_card"`
	IBAN       bool `yaml:"iban"`
	Email      bool `yaml:"email"`
	Phone      bool `yaml:"phone"`
	SSN        bool `yaml:"ssn"`
	IPAddress  bool `yaml:"ip_address"`
	PersonName bool `yaml:"person_name"`
	APISecrets bool `yaml:"api_secrets"`
}

type GovernanceConfig struct {
	Enabled                   bool   `yaml:"enabled"`
	EnforceArticle5Block      bool   `yaml:"enforce_article5_block"`
	LogAnnexIIIHighRisk       bool   `yaml:"log_annex_iii_high_risk"`
	RequireHumanOversightFlag bool   `yaml:"require_human_oversight_flag"`
	DefaultApplicationID      string `yaml:"default_application_id"`
}

type SecurityConfig struct {
	ScanPromptInjection  bool `yaml:"scan_prompt_injection"`
	BlockPromptInjection bool `yaml:"block_prompt_injection"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			UpstreamOpenAI:    "https://api.openai.com",
			UpstreamAnthropic: "https://api.anthropic.com",
			DBPath:            "./guardmesh.db",
		},
		PII: PIIConfig{
			Enabled:         true,
			MaskMode:        "token",
			TokenTTLSeconds: 300,
			Detectors: DetectorsConfig{
				CreditCard: true,
				IBAN:       true,
				Email:      true,
				Phone:      true,
				SSN:        true,
				IPAddress:  true,
				PersonName: true,
				APISecrets: true,
			},
		},
		Governance: GovernanceConfig{
			Enabled:                   true,
			EnforceArticle5Block:      true,
			LogAnnexIIIHighRisk:       true,
			RequireHumanOversightFlag: true,
			DefaultApplicationID:      "default-app",
		},
		Security: SecurityConfig{
			ScanPromptInjection:  true,
			BlockPromptInjection: false,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config yaml: %w", err)
	}

	return cfg, nil
}
