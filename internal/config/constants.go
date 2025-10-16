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
		ExtraArgs:      []string{"--disable-ui"},
	},
	Logging: LoggingConfig{
		Level:  "info",
		Format: "json",
	},
	Decision: DecisionConfig{
		BlockOnFindings:   true,
		SeverityThreshold: "high",
	},
}

// GetDefaultConfigDir returns the default configuration directory
func GetDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".hook-vault-radar"
	}
	return filepath.Join(home, ".hook-vault-radar")
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	return filepath.Join(GetDefaultConfigDir(), "config.yaml")
}
