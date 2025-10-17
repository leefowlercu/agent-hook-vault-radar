package strategies

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

// Helper function to create a test RemediationInput
func createTestInput() types.RemediationInput {
	return types.RemediationInput{
		ScanResults: types.ScanResults{
			HasFindings: true,
			Findings: []types.Finding{
				{
					Severity:    "info",
					Type:        "aws_access_key_id",
					Location:    "test-file.txt",
					Description: "AWS access key ID",
				},
				{
					Severity:    "high",
					Type:        "github_token",
					Location:    "test-file.txt",
					Description: "GitHub personal access token",
				},
			},
		},
		HookInput: types.HookInput{
			Framework: "claude",
			HookType:  "UserPromptSubmit",
			RawData: map[string]any{
				"session_id": "test-session-123",
			},
		},
		Decision: types.Decision{
			Block:  true,
			Reason: "Security findings detected",
		},
		Timestamp: time.Date(2025, 10, 16, 14, 30, 45, 0, time.UTC),
		Framework: "claude",
	}
}

func TestNewLogStrategy_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		cfg    config.StrategyConfig
		expect error
	}{
		{
			name: "valid json config",
			cfg: config.StrategyConfig{
				Type: "log",
				Config: map[string]any{
					"log_file": "/tmp/test.log",
					"format":   "json",
				},
			},
			expect: nil,
		},
		{
			name: "valid text config",
			cfg: config.StrategyConfig{
				Type: "log",
				Config: map[string]any{
					"log_file": "/tmp/test.log",
					"format":   "text",
				},
			},
			expect: nil,
		},
		{
			name: "default format",
			cfg: config.StrategyConfig{
				Type: "log",
				Config: map[string]any{
					"log_file": "/tmp/test.log",
				},
			},
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := NewLogStrategy(tt.cfg)
			if err != nil && tt.expect == nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && tt.expect != nil {
				t.Errorf("expected error but got none")
			}
			if err == nil && strategy == nil {
				t.Error("expected strategy but got nil")
			}
		})
	}
}

func TestNewLogStrategy_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		cfg    config.StrategyConfig
		errMsg string
	}{
		{
			name: "missing log_file",
			cfg: config.StrategyConfig{
				Type: "log",
				Config: map[string]any{
					"format": "json",
				},
			},
			errMsg: "log_file is required",
		},
		{
			name: "empty log_file",
			cfg: config.StrategyConfig{
				Type: "log",
				Config: map[string]any{
					"log_file": "",
					"format":   "json",
				},
			},
			errMsg: "log_file is required",
		},
		{
			name: "invalid format",
			cfg: config.StrategyConfig{
				Type: "log",
				Config: map[string]any{
					"log_file": "/tmp/test.log",
					"format":   "xml",
				},
			},
			errMsg: "format must be 'json' or 'text'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLogStrategy(tt.cfg)
			if err == nil {
				t.Fatal("expected error but got none")
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
			}
		})
	}
}

func TestLogStrategy_GetType(t *testing.T) {
	strategy := &LogStrategy{
		logFile: "/tmp/test.log",
		format:  "json",
	}

	if got := strategy.GetType(); got != "log" {
		t.Errorf("GetType() = %q, want %q", got, "log")
	}
}

func TestLogStrategy_Validate(t *testing.T) {
	tests := []struct {
		name      string
		strategy  *LogStrategy
		expectErr bool
	}{
		{
			name: "valid json",
			strategy: &LogStrategy{
				logFile: "/tmp/test.log",
				format:  "json",
			},
			expectErr: false,
		},
		{
			name: "valid text",
			strategy: &LogStrategy{
				logFile: "/tmp/test.log",
				format:  "text",
			},
			expectErr: false,
		},
		{
			name: "empty log file",
			strategy: &LogStrategy{
				logFile: "",
				format:  "json",
			},
			expectErr: true,
		},
		{
			name: "invalid format",
			strategy: &LogStrategy{
				logFile: "/tmp/test.log",
				format:  "yaml",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.strategy.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("Validate() error = %v, expectErr = %v", err, tt.expectErr)
			}
		})
	}
}

func TestLogStrategy_ExecuteJSON(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	strategy := &LogStrategy{
		logFile: logFile,
		format:  "json",
	}

	input := createTestInput()
	ctx := context.Background()

	result := strategy.Execute(ctx, input)

	// Check result
	if !result.Success {
		t.Fatalf("Execute() failed: %v", result.Error)
	}
	if result.StrategyType != "log" {
		t.Errorf("StrategyType = %q, want %q", result.StrategyType, "log")
	}
	if !strings.Contains(result.Message, "Logged 2 findings") {
		t.Errorf("Message = %q, want to contain 'Logged 2 findings'", result.Message)
	}

	// Check file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatal("log file was not created")
	}

	// Read and parse JSON
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var logEntry map[string]any
	if err := json.Unmarshal(data, &logEntry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Verify JSON structure
	if logEntry["framework"] != "claude" {
		t.Errorf("framework = %v, want 'claude'", logEntry["framework"])
	}
	if logEntry["session_id"] != "test-session-123" {
		t.Errorf("session_id = %v, want 'test-session-123'", logEntry["session_id"])
	}
	if logEntry["blocked"] != true {
		t.Errorf("blocked = %v, want true", logEntry["blocked"])
	}
	if logEntry["finding_count"] != float64(2) {
		t.Errorf("finding_count = %v, want 2", logEntry["finding_count"])
	}
}

func TestLogStrategy_ExecuteText(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	strategy := &LogStrategy{
		logFile: logFile,
		format:  "text",
	}

	input := createTestInput()
	ctx := context.Background()

	result := strategy.Execute(ctx, input)

	// Check result
	if !result.Success {
		t.Fatalf("Execute() failed: %v", result.Error)
	}

	// Read log file
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)

	// Verify text format
	if !strings.Contains(content, "[2025-10-16 14:30:45]") {
		t.Error("missing timestamp")
	}
	if !strings.Contains(content, "Framework: claude") {
		t.Error("missing framework")
	}
	if !strings.Contains(content, "Session: test-session-123") {
		t.Error("missing session ID")
	}
	if !strings.Contains(content, "Findings: 2") {
		t.Error("missing finding count")
	}
	if !strings.Contains(content, "Blocked: true") {
		t.Error("missing blocked status")
	}
	if !strings.Contains(content, "[INFO] aws_access_key_id") {
		t.Error("missing first finding")
	}
	if !strings.Contains(content, "[HIGH] github_token") {
		t.Error("missing second finding")
	}
}

func TestLogStrategy_FileCreation(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "nested", "dir", "test.log")

	strategy := &LogStrategy{
		logFile: logFile,
		format:  "json",
	}

	input := createTestInput()
	ctx := context.Background()

	result := strategy.Execute(ctx, input)

	// Check result
	if !result.Success {
		t.Fatalf("Execute() failed: %v", result.Error)
	}

	// Verify nested directories were created
	if _, err := os.Stat(filepath.Dir(logFile)); os.IsNotExist(err) {
		t.Error("parent directories were not created")
	}

	// Verify file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("log file was not created")
	}
}

func TestLogStrategy_AppendMode(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	strategy := &LogStrategy{
		logFile: logFile,
		format:  "json",
	}

	input := createTestInput()
	ctx := context.Background()

	// Execute twice
	result1 := strategy.Execute(ctx, input)
	if !result1.Success {
		t.Fatalf("first Execute() failed: %v", result1.Error)
	}

	result2 := strategy.Execute(ctx, input)
	if !result2.Success {
		t.Fatalf("second Execute() failed: %v", result2.Error)
	}

	// Read file
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Count number of JSON lines
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Verify both lines are valid JSON
	for i, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i+1, err)
		}
	}
}

func TestLogStrategy_ContextCancellation(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	strategy := &LogStrategy{
		logFile: logFile,
		format:  "json",
	}

	input := createTestInput()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := strategy.Execute(ctx, input)

	// Should fail due to cancellation
	if result.Success {
		t.Error("Execute() succeeded with cancelled context, expected failure")
	}
	if !strings.Contains(result.Message, "cancelled") {
		t.Errorf("expected cancellation message, got: %s", result.Message)
	}
}

func TestLogStrategy_PathExpansion(t *testing.T) {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home directory: %v", err)
	}

	// Create temp file in home directory
	tmpFile := filepath.Join(home, ".test-log-strategy-"+time.Now().Format("20060102150405")+".log")
	defer os.Remove(tmpFile) // Clean up

	// Use ~ in path
	tildeFile := "~/.test-log-strategy-" + time.Now().Format("20060102150405") + ".log"

	strategy := &LogStrategy{
		logFile: tildeFile,
		format:  "json",
	}

	input := createTestInput()
	ctx := context.Background()

	result := strategy.Execute(ctx, input)

	// Check result
	if !result.Success {
		t.Fatalf("Execute() failed: %v", result.Error)
	}

	// Verify file was created in home directory
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("log file was not created with ~ expansion")
	}
}

func TestLogStrategy_WriteError(t *testing.T) {
	// Try to write to a directory (should fail)
	tmpDir := t.TempDir()

	strategy := &LogStrategy{
		logFile: tmpDir, // Directory, not a file
		format:  "json",
	}

	input := createTestInput()
	ctx := context.Background()

	result := strategy.Execute(ctx, input)

	// Should fail
	if result.Success {
		t.Error("Execute() succeeded when writing to directory, expected failure")
	}
	if result.Error == nil {
		t.Error("expected error but got nil")
	}
}

func TestLogStrategy_FormatJSON(t *testing.T) {
	strategy := &LogStrategy{
		logFile: "/tmp/test.log",
		format:  "json",
	}

	input := createTestInput()

	content, err := strategy.formatJSON(input)
	if err != nil {
		t.Fatalf("formatJSON() failed: %v", err)
	}

	// Parse JSON
	var entry map[string]any
	if err := json.Unmarshal([]byte(content), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify fields
	if entry["framework"] != "claude" {
		t.Errorf("framework = %v, want 'claude'", entry["framework"])
	}
	if entry["finding_count"] != float64(2) {
		t.Errorf("finding_count = %v, want 2", entry["finding_count"])
	}
}

func TestLogStrategy_FormatText(t *testing.T) {
	strategy := &LogStrategy{
		logFile: "/tmp/test.log",
		format:  "text",
	}

	input := createTestInput()

	content, err := strategy.formatText(input)
	if err != nil {
		t.Fatalf("formatText() failed: %v", err)
	}

	// Verify content
	if !strings.Contains(content, "Framework: claude") {
		t.Error("missing framework in text format")
	}
	if !strings.Contains(content, "Findings: 2") {
		t.Error("missing finding count in text format")
	}
	if !strings.Contains(content, "[INFO] aws_access_key_id") {
		t.Error("missing first finding in text format")
	}
}
