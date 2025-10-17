package config

// Config represents the application configuration
type Config struct {
	VaultRadar  VaultRadarConfig   `mapstructure:"vault_radar" yaml:"vault_radar"`
	Logging     LoggingConfig      `mapstructure:"logging" yaml:"logging"`
	Decision    DecisionConfig     `mapstructure:"decision" yaml:"decision"`
	Remediation RemediationConfig  `mapstructure:"remediation" yaml:"remediation"`
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
	Level   string `mapstructure:"level" yaml:"level"`
	Format  string `mapstructure:"format" yaml:"format"`
	LogFile string `mapstructure:"log_file" yaml:"log_file"` // Optional file path for logging (empty = stderr only)
}

// DecisionConfig contains configuration for decision-making logic
type DecisionConfig struct {
	BlockOnFindings   bool   `mapstructure:"block_on_findings" yaml:"block_on_findings"`
	SeverityThreshold string `mapstructure:"severity_threshold" yaml:"severity_threshold"`
}

// RemediationConfig contains configuration for remediation actions
type RemediationConfig struct {
	Enabled        bool             `mapstructure:"enabled" yaml:"enabled"`
	TimeoutSeconds int              `mapstructure:"timeout_seconds" yaml:"timeout_seconds"`
	Protocols      []ProtocolConfig `mapstructure:"protocols" yaml:"protocols"`
}

// ProtocolConfig defines a remediation protocol with triggers and strategies
type ProtocolConfig struct {
	Name       string           `mapstructure:"name" yaml:"name"`
	Triggers   TriggerConfig    `mapstructure:"triggers" yaml:"triggers"`
	Strategies []StrategyConfig `mapstructure:"strategies" yaml:"strategies"`
}

// TriggerConfig defines when a protocol should execute
type TriggerConfig struct {
	OnBlock           bool     `mapstructure:"on_block" yaml:"on_block"`
	OnFindings        bool     `mapstructure:"on_findings" yaml:"on_findings"`
	SeverityThreshold string   `mapstructure:"severity_threshold" yaml:"severity_threshold"`
	FindingTypes      []string `mapstructure:"finding_types" yaml:"finding_types"`
}

// StrategyConfig defines a remediation strategy configuration
type StrategyConfig struct {
	Type   string         `mapstructure:"type" yaml:"type"`
	Config map[string]any `mapstructure:"config" yaml:"config"`
}
