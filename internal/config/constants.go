package config

import (
	"os"
	"path/filepath"
)

// DefaultConfig provides default configuration values
var DefaultConfig = Config{
	VaultRadar: VaultRadarConfig{
		Command:        "vault-radar",
		ScanCommand:    "scan file",
		TimeoutSeconds: 30,
		ExtraArgs:      []string{},
	},
	Logging: LoggingConfig{
		Level:   "info",
		Format:  "json",
		LogFile: "~/.agent-hooks/vault-radar/logs/hook.log", // File-only logging (no stderr)
	},
	Decision: DecisionConfig{
		BlockOnFindings:   true,
		SeverityThreshold: "medium",
	},
	Remediation: RemediationConfig{
		Enabled:        false,              // Disabled by default, opt-in feature
		TimeoutSeconds: 10,                 // 10 second timeout for all remediation strategies
		Protocols:      []ProtocolConfig{}, // No default protocols, must be configured
	},
}

// GetDefaultConfigDir returns the default configuration directory
func GetDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".agent-hooks/vault-radar"
	}
	return filepath.Join(home, ".agent-hooks/vault-radar")
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	return filepath.Join(GetDefaultConfigDir(), "config.yaml")
}
