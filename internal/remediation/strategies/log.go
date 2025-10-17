package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// LogStrategy implements a remediation strategy that logs finding details to a file
type LogStrategy struct {
	logFile string // Path to log file (supports ~ expansion)
	format  string // "json" or "text"
}

// NewLogStrategy creates a new log strategy from configuration
func NewLogStrategy(cfg config.StrategyConfig) (*LogStrategy, error) {
	logFile, ok := cfg.Config["log_file"].(string)
	if !ok || logFile == "" {
		return nil, fmt.Errorf("log_file is required")
	}

	format, ok := cfg.Config["format"].(string)
	if !ok || format == "" {
		format = "json" // Default to JSON
	}

	strategy := &LogStrategy{
		logFile: logFile,
		format:  format,
	}

	if err := strategy.Validate(); err != nil {
		return nil, err
	}

	return strategy, nil
}

// Execute writes finding details to the configured log file
func (s *LogStrategy) Execute(ctx context.Context, input types.RemediationInput) types.RemediationResult {
	// Check for context cancellation before starting
	select {
	case <-ctx.Done():
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      "Log operation cancelled",
			Error:        ctx.Err(),
		}
	default:
	}

	// Expand file path
	logPath, err := s.expandPath(s.logFile)
	if err != nil {
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      fmt.Sprintf("Failed to expand log path: %v", err),
			Error:        err,
		}
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      fmt.Sprintf("Failed to create log directory: %v", err),
			Error:        err,
		}
	}

	// Open file in append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      fmt.Sprintf("Failed to open log file: %v", err),
			Error:        err,
		}
	}
	defer file.Close()

	// Format and write content
	var content string
	var formatErr error

	switch s.format {
	case "json":
		content, formatErr = s.formatJSON(input)
	case "text":
		content, formatErr = s.formatText(input)
	default:
		formatErr = fmt.Errorf("unsupported format: %s", s.format)
	}

	if formatErr != nil {
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      fmt.Sprintf("Failed to format log content: %v", formatErr),
			Error:        formatErr,
		}
	}

	// Check context again before writing
	select {
	case <-ctx.Done():
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      "Log operation cancelled before write",
			Error:        ctx.Err(),
		}
	default:
	}

	// Write to file
	if _, err := file.WriteString(content + "\n"); err != nil {
		return types.RemediationResult{
			StrategyType: s.GetType(),
			Success:      false,
			Message:      fmt.Sprintf("Failed to write to log file: %v", err),
			Error:        err,
		}
	}

	// Build success message
	findingCount := len(input.ScanResults.Findings)
	var message string
	if findingCount == 1 {
		message = fmt.Sprintf("Logged 1 finding to %s", filepath.Base(logPath))
	} else {
		message = fmt.Sprintf("Logged %d findings to %s", findingCount, filepath.Base(logPath))
	}

	return types.RemediationResult{
		StrategyType: s.GetType(),
		Success:      true,
		Message:      message,
		Metadata: map[string]any{
			"log_file":      logPath,
			"format":        s.format,
			"finding_count": findingCount,
		},
	}
}

// GetType returns the strategy type identifier
func (s *LogStrategy) GetType() string {
	return "log"
}

// Validate checks if the strategy configuration is valid
func (s *LogStrategy) Validate() error {
	if s.logFile == "" {
		return fmt.Errorf("log_file cannot be empty")
	}

	if s.format != "json" && s.format != "text" {
		return fmt.Errorf("format must be 'json' or 'text', got: %s", s.format)
	}

	return nil
}

// expandPath expands ~ to the user's home directory
func (s *LogStrategy) expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

// formatJSON formats the log entry as JSON
func (s *LogStrategy) formatJSON(input types.RemediationInput) (string, error) {
	// Extract session ID from hook input if available
	sessionID := ""
	if sid, ok := input.HookInput.RawData["session_id"].(string); ok {
		sessionID = sid
	}

	// Build JSON structure
	logEntry := map[string]any{
		"timestamp":     input.Timestamp.Format(time.RFC3339),
		"framework":     input.Framework,
		"session_id":    sessionID,
		"blocked":       input.Decision.Block,
		"finding_count": len(input.ScanResults.Findings),
		"findings":      input.ScanResults.Findings,
	}

	// Marshal to JSON
	data, err := json.Marshal(logEntry)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// formatText formats the log entry as human-readable text
func (s *LogStrategy) formatText(input types.RemediationInput) (string, error) {
	var sb strings.Builder

	// Extract session ID from hook input if available
	sessionID := ""
	if sid, ok := input.HookInput.RawData["session_id"].(string); ok {
		sessionID = sid
	}

	// Build header line
	timestamp := input.Timestamp.Format("2006-01-02 15:04:05")
	findingCount := len(input.ScanResults.Findings)
	blocked := "false"
	if input.Decision.Block {
		blocked = "true"
	}

	sb.WriteString(fmt.Sprintf("[%s] Framework: %s | Session: %s | Findings: %d | Blocked: %s",
		timestamp, input.Framework, sessionID, findingCount, blocked))

	// Add findings details
	for _, finding := range input.ScanResults.Findings {
		sb.WriteString("\n  - [")
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
	}

	return sb.String(), nil
}
