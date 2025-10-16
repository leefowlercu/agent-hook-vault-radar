package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// InitConfig initializes the configuration using Viper
func InitConfig() error {
	// Load .env file if it exists (fail silently if not found)
	loadEnvFiles()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(GetDefaultConfigDir())
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("framework", DefaultConfig.Framework)
	viper.SetDefault("vault_radar.command", DefaultConfig.VaultRadar.Command)
	viper.SetDefault("vault_radar.scan_command", DefaultConfig.VaultRadar.ScanCommand)
	viper.SetDefault("vault_radar.timeout_seconds", DefaultConfig.VaultRadar.TimeoutSeconds)
	viper.SetDefault("vault_radar.extra_args", DefaultConfig.VaultRadar.ExtraArgs)
	viper.SetDefault("logging.level", DefaultConfig.Logging.Level)
	viper.SetDefault("logging.format", DefaultConfig.Logging.Format)
	viper.SetDefault("decision.block_on_findings", DefaultConfig.Decision.BlockOnFindings)
	viper.SetDefault("decision.severity_threshold", DefaultConfig.Decision.SeverityThreshold)

	// Enable environment variable overrides
	viper.SetEnvPrefix("HOOK_VAULT_RADAR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file (it's okay if it doesn't exist)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config; %w", err)
		}
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config; %w", err)
	}

	return &cfg, nil
}

// loadEnvFiles loads environment variables from .env files
// It tries multiple locations and fails silently if files don't exist
func loadEnvFiles() {
	// Try to load .env files from multiple locations
	locations := []string{
		".env", // Current directory
		filepath.Join(GetDefaultConfigDir(), ".env"), // Config directory (~/.hook-vault-radar/.env)
	}

	// Also try .env.local for local overrides
	localLocations := []string{
		".env.local",
		filepath.Join(GetDefaultConfigDir(), ".env.local"),
	}

	// Load .env files first
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			_ = godotenv.Load(location) // Fail silently
		}
	}

	// Load .env.local files (override .env)
	for _, location := range localLocations {
		if _, err := os.Stat(location); err == nil {
			_ = godotenv.Load(location) // Fail silently
		}
	}
}
