package scanner

import (
	"context"

	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// Scanner defines the interface for security scanners
type Scanner interface {
	// Scan scans content for secrets and sensitive data
	Scan(ctx context.Context, content types.ScanContent) (types.ScanResults, error)

	// GetName returns the scanner name
	GetName() string
}
