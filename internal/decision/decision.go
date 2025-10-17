package decision

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// Engine makes decisions based on scan results and configuration
type Engine struct {
	cfg *config.Config
}

// NewEngine creates a new decision engine
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		cfg: cfg,
	}
}

// Evaluate evaluates scan results and produces a decision
func (e *Engine) Evaluate(ctx context.Context, results types.ScanResults) (types.Decision, error) {
	decision := types.Decision{
		Block:    false,
		Metadata: make(map[string]any),
	}

	// If there was an error during scanning, decide based on fail-open/fail-closed policy
	if results.Error != nil {
		decision.Metadata["scan_error"] = results.Error.Error()
		// Currently fail-open (allow on error), but this could be configurable
		return decision, nil
	}

	if !results.HasFindings {
		return decision, nil
	}

	// Filter findings by severity threshold
	relevantFindings := e.filterBySeverity(results.Findings)

	if len(relevantFindings) == 0 {
		// No findings meet the threshold
		decision.Metadata["filtered_findings"] = results.Findings
		return decision, nil
	}

	// Block if configured to do so and we have relevant findings
	if e.cfg.Decision.BlockOnFindings {
		decision.Block = true
		decision.Reason = e.buildReasonMessage(relevantFindings)
		decision.Metadata["findings"] = relevantFindings
		decision.Metadata["finding_count"] = len(relevantFindings)
	}

	return decision, nil
}

// filterBySeverity filters findings based on the configured severity threshold
func (e *Engine) filterBySeverity(findings []types.Finding) []types.Finding {
	threshold := e.getSeverityLevel(e.cfg.Decision.SeverityThreshold)
	filtered := []types.Finding{}

	for _, finding := range findings {
		findingSeverity := e.getSeverityLevel(finding.Severity)
		if findingSeverity >= threshold {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}

// getSeverityLevel converts severity string to numeric level for comparison
func (e *Engine) getSeverityLevel(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium", "info": // vault-radar uses "info" for many real secrets
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// buildReasonMessage creates a human-readable explanation of why the action was blocked
func (e *Engine) buildReasonMessage(findings []types.Finding) string {
	if len(findings) == 0 {
		return "Security scan completed with no findings"
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString("Vault Radar detected ")

	if len(findings) == 1 {
		sb.WriteString("1 security finding:\n\n")
	} else {
		sb.WriteString(strconv.Itoa(len(findings)))
		sb.WriteString(" security findings:\n\n")
	}

	for i, finding := range findings {
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(". [")
		sb.WriteString(strings.ToUpper(finding.Severity))
		sb.WriteString("] ")
		sb.WriteString(finding.Type)

		if finding.Description != "" {
			sb.WriteString(": ")
			sb.WriteString(finding.Description)
		}

		if finding.Location != "" {
			sb.WriteString(" (")
			sb.WriteString(finding.Location)
			sb.WriteString(")")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("\nPlease remove or redact sensitive information before proceeding.")

	return sb.String()
}

// EnrichWithRemediation appends remediation results to the decision reason
func EnrichWithRemediation(decision *types.Decision, results types.RemediationResults) {
	if !results.Executed || len(results.Results) == 0 {
		return
	}

	summary := buildRemediationSummary(results)
	if decision.Reason != "" {
		decision.Reason += "\n\n" + summary
	} else {
		decision.Reason = summary
	}
}

// buildRemediationSummary creates a formatted summary of remediation results
func buildRemediationSummary(results types.RemediationResults) string {
	var sb strings.Builder

	// Header with strategy count and total duration
	sb.WriteString("Remediation actions taken (")
	sb.WriteString(strconv.Itoa(len(results.Results)))
	if len(results.Results) == 1 {
		sb.WriteString(" strategy, ")
	} else {
		sb.WriteString(" strategies, ")
	}
	sb.WriteString(formatDuration(results.TotalDuration))
	sb.WriteString(" total):")

	// Individual strategy results
	for _, result := range results.Results {
		sb.WriteString("\n  ")

		// Success/failure indicator
		if result.Success {
			sb.WriteString("✓ ") // U+2713 check mark
		} else {
			sb.WriteString("✗ ") // U+2717 ballot x
		}

		// Strategy message
		sb.WriteString(result.Message)

		// Duration
		sb.WriteString(" (")
		sb.WriteString(formatDuration(result.Duration))
		sb.WriteString(")")
	}

	return sb.String()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()

	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}

	seconds := float64(ms) / 1000.0
	return fmt.Sprintf("%.1fs", seconds)
}
