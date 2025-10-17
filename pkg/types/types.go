package types

import "time"

// ScanContent represents content to be scanned by Vault Radar
type ScanContent struct {
	Type     string            // "text", "file", "directory"
	Content  string            // The content to scan
	Metadata map[string]string // Additional context
}

// Finding represents a single security finding from a scan
type Finding struct {
	Severity    string // "high", "medium", "low"
	Type        string // "secret", "credential", "api_key", etc.
	Location    string // Where the finding was detected
	Description string // Human-readable description
}

// ScanResults contains the results of a Vault Radar scan
type ScanResults struct {
	HasFindings  bool
	Findings     []Finding
	ScanDuration time.Duration
	Error        error
}

// Decision represents the hook's decision on whether to proceed or block
type Decision struct {
	Block    bool           // Whether to block the action
	Reason   string         // Human-readable explanation
	Metadata map[string]any // Additional metadata for the hook framework
}

// HookInput represents parsed input from a hook framework
type HookInput struct {
	Framework string         // Framework name (e.g., "claude")
	HookType  string         // Hook type (e.g., "UserPromptSubmit")
	RawData   map[string]any // Raw JSON data from stdin
}

// RemediationInput contains all context needed for remediation strategies
type RemediationInput struct {
	ScanResults ScanResults  // Complete scan results (includes findings)
	HookInput   HookInput    // Original hook input
	Decision    Decision     // Decision made by the decision engine
	Timestamp   time.Time    // When the remediation is being executed
	Framework   string       // Framework name for context
}

// RemediationResult represents the result of executing a single remediation strategy
type RemediationResult struct {
	StrategyType string         // Type of strategy that executed (e.g., "log", "webhook")
	Success      bool           // Whether the strategy executed successfully
	Message      string         // User-facing summary message
	Duration     time.Duration  // How long the strategy took to execute
	Metadata     map[string]any // Additional metadata from the strategy
	Error        error          // Error if the strategy failed
}

// RemediationResults represents the aggregate results from executing a remediation protocol
type RemediationResults struct {
	Executed      bool                // Whether remediation was executed
	Results       []RemediationResult // Individual strategy results
	TotalDuration time.Duration       // Total time for all strategies
	ProtocolName  string              // Name of the protocol that was executed
}
