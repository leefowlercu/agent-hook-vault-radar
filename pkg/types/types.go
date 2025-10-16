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
	ExitCode int            // Exit code to return
}

// HookInput represents parsed input from a hook framework
type HookInput struct {
	Framework string         // Framework name (e.g., "claude")
	HookType  string         // Hook type (e.g., "UserPromptSubmit")
	RawData   map[string]any // Raw JSON data from stdin
}
