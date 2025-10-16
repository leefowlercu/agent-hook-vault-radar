package framework

import (
	"context"
	"io"

	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// HookFramework defines the interface for hook framework implementations
type HookFramework interface {
	// ParseInput reads and parses hook data from stdin
	ParseInput(reader io.Reader) (types.HookInput, error)

	// FormatOutput formats a decision as JSON for the framework
	FormatOutput(decision types.Decision, input types.HookInput) ([]byte, error)

	// GetName returns the framework name
	GetName() string
}

// HookHandler defines the interface for specific hook type handlers
type HookHandler interface {
	// ExtractContent extracts scannable content from hook input
	ExtractContent(ctx context.Context, input types.HookInput) (types.ScanContent, error)

	// MakeDecision creates a decision based on scan results
	MakeDecision(ctx context.Context, results types.ScanResults, input types.HookInput) (types.Decision, error)

	// GetType returns the hook type name
	GetType() string

	// CanHandle returns true if this handler can process the given hook input
	CanHandle(input types.HookInput) bool
}
