package config

import (
	"os"
	"path/filepath"
)

// DefaultConfig provides default configuration values
var DefaultConfig = Config{
	Framework: "claude",
	VaultRadar: VaultRadarConfig{
		Command:        "vault-radar",
		ScanCommand:    "scan file",
		TimeoutSeconds: 30,
		ExtraArgs:      []string{},
	},
	Logging: LoggingConfig{
		Level:   "info",
		Format:  "json",
		LogFile: "", // Empty = stderr only, set path to enable file logging
	},
	Decision: DecisionConfig{
		BlockOnFindings:   true,
		SeverityThreshold: "medium",
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
