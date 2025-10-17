package remediation

import (
	"strings"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// Protocol represents a remediation protocol with triggers and strategies
type Protocol struct {
	Name       string
	Triggers   config.TriggerConfig
	Strategies []config.StrategyConfig
}

// NewProtocol creates a new protocol from configuration
func NewProtocol(cfg config.ProtocolConfig) *Protocol {
	return &Protocol{
		Name:       cfg.Name,
		Triggers:   cfg.Triggers,
		Strategies: cfg.Strategies,
	}
}

// ShouldExecute determines if this protocol's triggers match the current state
func (p *Protocol) ShouldExecute(input types.RemediationInput) bool {
	// Check on_block trigger
	if p.Triggers.OnBlock && !input.Decision.Block {
		return false
	}

	// Check on_findings trigger
	if p.Triggers.OnFindings && !input.ScanResults.HasFindings {
		return false
	}

	// If both on_block and on_findings are false, protocol never executes
	if !p.Triggers.OnBlock && !p.Triggers.OnFindings {
		return false
	}

	// Check severity threshold if specified
	if p.Triggers.SeverityThreshold != "" && input.ScanResults.HasFindings {
		if !p.matchesSeverityThreshold(input.ScanResults.Findings, p.Triggers.SeverityThreshold) {
			return false
		}
	}

	// Check finding types if specified
	if len(p.Triggers.FindingTypes) > 0 && input.ScanResults.HasFindings {
		if !p.matchesFindingTypes(input.ScanResults.Findings, p.Triggers.FindingTypes) {
			return false
		}
	}

	return true
}

// matchesSeverityThreshold checks if any finding meets or exceeds the severity threshold
func (p *Protocol) matchesSeverityThreshold(findings []types.Finding, threshold string) bool {
	thresholdLevel := getSeverityLevel(threshold)

	for _, finding := range findings {
		findingLevel := getSeverityLevel(finding.Severity)
		if findingLevel >= thresholdLevel {
			return true
		}
	}

	return false
}

// matchesFindingTypes checks if any finding matches the specified type patterns
func (p *Protocol) matchesFindingTypes(findings []types.Finding, patterns []string) bool {
	for _, finding := range findings {
		for _, pattern := range patterns {
			if matchesPattern(finding.Type, pattern) {
				return true
			}
		}
	}

	return false
}

// getSeverityLevel converts severity string to numeric level for comparison
func getSeverityLevel(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium", "info":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// matchesPattern checks if a finding type matches a pattern (supports wildcards)
func matchesPattern(findingType string, pattern string) bool {
	// Simple wildcard matching: * matches any characters
	// Example: "aws_*" matches "aws_access_key_id", "aws_secret_key", etc.

	if pattern == "*" {
		return true
	}

	if !strings.Contains(pattern, "*") {
		// No wildcard, exact match
		return findingType == pattern
	}

	// Split pattern by * and check each part
	parts := strings.Split(pattern, "*")

	// Check prefix
	if len(parts[0]) > 0 && !strings.HasPrefix(findingType, parts[0]) {
		return false
	}

	// Check suffix
	if len(parts) > 1 && len(parts[len(parts)-1]) > 0 {
		if !strings.HasSuffix(findingType, parts[len(parts)-1]) {
			return false
		}
	}

	// Check middle parts
	currentPos := len(parts[0])
	for i := 1; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}
		idx := strings.Index(findingType[currentPos:], part)
		if idx == -1 {
			return false
		}
		currentPos += idx + len(part)
	}

	return true
}
