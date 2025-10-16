package config

// Config represents the application configuration
type Config struct {
	Framework  string           `mapstructure:"framework" yaml:"framework"`
	VaultRadar VaultRadarConfig `mapstructure:"vault_radar" yaml:"vault_radar"`
	Logging    LoggingConfig    `mapstructure:"logging" yaml:"logging"`
	Decision   DecisionConfig   `mapstructure:"decision" yaml:"decision"`
}

// VaultRadarConfig contains configuration for the Vault Radar CLI
type VaultRadarConfig struct {
	Command        string   `mapstructure:"command" yaml:"command"`
	ScanCommand    string   `mapstructure:"scan_command" yaml:"scan_command"`
	TimeoutSeconds int      `mapstructure:"timeout_seconds" yaml:"timeout_seconds"`
	ExtraArgs      []string `mapstructure:"extra_args" yaml:"extra_args"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" yaml:"level"`
	Format string `mapstructure:"format" yaml:"format"`
}

// DecisionConfig contains configuration for decision-making logic
type DecisionConfig struct {
	BlockOnFindings   bool   `mapstructure:"block_on_findings" yaml:"block_on_findings"`
	SeverityThreshold string `mapstructure:"severity_threshold" yaml:"severity_threshold"`
}
