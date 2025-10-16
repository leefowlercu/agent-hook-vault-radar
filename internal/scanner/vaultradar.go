package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

const scannerName = "vault-radar"

// VaultRadarScanner implements the Scanner interface using Vault Radar CLI
type VaultRadarScanner struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewVaultRadarScanner creates a new Vault Radar scanner instance
func NewVaultRadarScanner(cfg *config.Config, logger *slog.Logger) *VaultRadarScanner {
	return &VaultRadarScanner{
		cfg:    cfg,
		logger: logger,
	}
}

// Scan executes vault-radar to scan the provided content
func (s *VaultRadarScanner) Scan(ctx context.Context, content types.ScanContent) (types.ScanResults, error) {
	startTime := time.Now()

	results := types.ScanResults{
		HasFindings: false,
		Findings:    []types.Finding{},
	}

	// Create a temporary directory for scanning
	tempDir, err := os.MkdirTemp("", "vault-radar-scan-*")
	if err != nil {
		results.Error = fmt.Errorf("failed to create temp directory; %w", err)
		return results, results.Error
	}
	defer os.RemoveAll(tempDir)

	// Write content to a temporary file
	tempFile := filepath.Join(tempDir, "scan-content.txt")
	if err := os.WriteFile(tempFile, []byte(content.Content), 0600); err != nil {
		results.Error = fmt.Errorf("failed to write temp file; %w", err)
		return results, results.Error
	}

	// Create output file for vault-radar results
	outputFile := filepath.Join(tempDir, "vault-radar-output.json")

	s.logger.Debug("created temporary file for scanning",
		"file", tempFile,
		"output_file", outputFile,
		"content_length", len(content.Content))

	// Build vault-radar command
	cmdArgs := s.buildCommandArgs(tempFile, outputFile)

	s.logger.Info("executing vault-radar",
		"command", s.cfg.VaultRadar.Command,
		"args", cmdArgs)

	// Create command with context for timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.VaultRadar.TimeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, s.cfg.VaultRadar.Command, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err = cmd.Run()

	results.ScanDuration = time.Since(startTime)

	s.logger.Debug("vault-radar execution completed",
		"duration", results.ScanDuration,
		"exit_code", cmd.ProcessState.ExitCode())

	// vault-radar returns non-zero exit code if secrets are found or on error
	if err != nil {
		// Check if it's a timeout
		if timeoutCtx.Err() == context.DeadlineExceeded {
			results.Error = fmt.Errorf("vault-radar scan timed out after %d seconds", s.cfg.VaultRadar.TimeoutSeconds)
			return results, results.Error
		}

		// Log stdout/stderr for debugging when vault-radar fails
		s.logger.Warn("vault-radar returned non-zero exit code",
			"error", err,
			"exit_code", cmd.ProcessState.ExitCode(),
			"stdout", stdout.String(),
			"stderr", stderr.String())
	}

	// Parse output file to extract findings
	findings, err := s.parseOutputFile(outputFile)
	if err != nil {
		s.logger.Warn("failed to parse vault-radar output file",
			"error", err,
			"output_file", outputFile)
		// Continue with empty findings rather than failing
		findings = []types.Finding{}
	}

	results.Findings = findings
	results.HasFindings = len(findings) > 0

	s.logger.Info("scan completed",
		"has_findings", results.HasFindings,
		"finding_count", len(findings),
		"duration", results.ScanDuration)

	return results, nil
}

// buildCommandArgs constructs the command arguments for vault-radar
func (s *VaultRadarScanner) buildCommandArgs(filePath, outputFile string) []string {
	// Start with the scan command (e.g., "scan file")
	args := strings.Fields(s.cfg.VaultRadar.ScanCommand)

	// Add the path flag with the file to scan
	args = append(args, "--path", filePath)

	// Add the required outfile flag
	args = append(args, "--outfile", outputFile)

	// Add format flag (must be two separate arguments)
	args = append(args, "--format", "json")

	// Add any extra arguments from config
	args = append(args, s.cfg.VaultRadar.ExtraArgs...)

	return args
}

// parseOutputFile parses vault-radar output file and extracts findings
// Vault-radar outputs newline-delimited JSON (NDJSON) - one JSON object per line
func (s *VaultRadarScanner) parseOutputFile(outputFile string) ([]types.Finding, error) {
	findings := []types.Finding{}

	// Check if output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return findings, fmt.Errorf("output file does not exist (vault-radar may have failed)")
	}

	// Read the output file
	data, err := os.ReadFile(outputFile)
	if err != nil {
		return findings, fmt.Errorf("failed to read output file; %w", err)
	}

	// Handle empty output (no secrets found)
	if len(data) == 0 {
		s.logger.Debug("vault-radar output file is empty (no secrets found)")
		return findings, nil
	}

	// Parse newline-delimited JSON (NDJSON)
	// Each line is a separate JSON object representing a finding
	lines := strings.Split(string(data), "\n")
	for lineNum, line := range lines {
		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var secretMap map[string]any
		if err := json.Unmarshal([]byte(line), &secretMap); err != nil {
			s.logger.Warn("failed to parse JSON line",
				"line_num", lineNum+1,
				"error", err)
			continue
		}

		finding := types.Finding{
			Severity: "high", // Default severity
			Type:     "secret",
		}

		if secretType, ok := secretMap["type"].(string); ok {
			finding.Type = secretType
		}

		if path, ok := secretMap["path"].(string); ok {
			finding.Location = path
		}

		if description, ok := secretMap["description"].(string); ok {
			finding.Description = description
		}

		if severity, ok := secretMap["severity"].(string); ok {
			finding.Severity = strings.ToLower(severity)
		}

		findings = append(findings, finding)
	}

	s.logger.Debug("parsed vault-radar output",
		"findings_count", len(findings),
		"output_size", len(data))

	return findings, nil
}

// GetName returns the scanner name
func (s *VaultRadarScanner) GetName() string {
	return scannerName
}
