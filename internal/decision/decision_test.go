package decision

import (
	"strings"
	"testing"
	"time"

	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

func TestEnrichWithRemediation(t *testing.T) {
	tests := []struct {
		name             string
		initialReason    string
		results          types.RemediationResults
		expectEnrichment bool
		expectContains   []string
	}{
		{
			name:          "successful remediation",
			initialReason: "Security findings detected",
			results: types.RemediationResults{
				Executed: true,
				Results: []types.RemediationResult{
					{
						StrategyType: "log",
						Success:      true,
						Message:      "Logged 2 findings to findings.log",
						Duration:     15 * time.Millisecond,
					},
				},
				TotalDuration: 15 * time.Millisecond,
				ProtocolName:  "default",
			},
			expectEnrichment: true,
			expectContains: []string{
				"Security findings detected",
				"Remediation actions taken",
				"✓ Logged 2 findings",
				"15ms",
			},
		},
		{
			name:          "not executed",
			initialReason: "Security findings detected",
			results: types.RemediationResults{
				Executed: false,
			},
			expectEnrichment: false,
			expectContains: []string{
				"Security findings detected",
			},
		},
		{
			name:          "no results",
			initialReason: "Security findings detected",
			results: types.RemediationResults{
				Executed: true,
				Results:  []types.RemediationResult{},
			},
			expectEnrichment: false,
			expectContains: []string{
				"Security findings detected",
			},
		},
		{
			name:          "empty initial reason",
			initialReason: "",
			results: types.RemediationResults{
				Executed: true,
				Results: []types.RemediationResult{
					{
						StrategyType: "log",
						Success:      true,
						Message:      "Logged findings",
						Duration:     10 * time.Millisecond,
					},
				},
				TotalDuration: 10 * time.Millisecond,
			},
			expectEnrichment: true,
			expectContains: []string{
				"Remediation actions taken",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := &types.Decision{
				Block:  true,
				Reason: tt.initialReason,
			}

			EnrichWithRemediation(decision, tt.results)

			// Check if enrichment happened as expected
			for _, expected := range tt.expectContains {
				if !strings.Contains(decision.Reason, expected) {
					t.Errorf("expected reason to contain %q, got: %s", expected, decision.Reason)
				}
			}

			// Verify enrichment happened or not
			if tt.expectEnrichment {
				if !strings.Contains(decision.Reason, "Remediation actions taken") {
					t.Error("expected enrichment but none found")
				}
			} else {
				if strings.Contains(decision.Reason, "Remediation actions taken") {
					t.Error("unexpected enrichment found")
				}
			}
		})
	}
}

func TestBuildRemediationSummary_Success(t *testing.T) {
	results := types.RemediationResults{
		Executed: true,
		Results: []types.RemediationResult{
			{
				StrategyType: "log",
				Success:      true,
				Message:      "Logged 2 findings to findings.log",
				Duration:     12 * time.Millisecond,
			},
			{
				StrategyType: "webhook",
				Success:      true,
				Message:      "Sent webhook notification to security system",
				Duration:     1200 * time.Millisecond,
			},
		},
		TotalDuration: 2300 * time.Millisecond,
		ProtocolName:  "default",
	}

	summary := buildRemediationSummary(results)

	// Check header
	if !strings.Contains(summary, "Remediation actions taken (2 strategies, 2.3s total):") {
		t.Errorf("unexpected header in summary: %s", summary)
	}

	// Check first strategy
	if !strings.Contains(summary, "✓ Logged 2 findings to findings.log (12ms)") {
		t.Errorf("missing first strategy in summary: %s", summary)
	}

	// Check second strategy
	if !strings.Contains(summary, "✓ Sent webhook notification to security system (1.2s)") {
		t.Errorf("missing second strategy in summary: %s", summary)
	}
}

func TestBuildRemediationSummary_Mixed(t *testing.T) {
	results := types.RemediationResults{
		Executed: true,
		Results: []types.RemediationResult{
			{
				StrategyType: "log",
				Success:      true,
				Message:      "Logged findings",
				Duration:     10 * time.Millisecond,
			},
			{
				StrategyType: "webhook",
				Success:      false,
				Message:      "Failed to send webhook: connection timeout",
				Duration:     5 * time.Second,
			},
			{
				StrategyType: "vault",
				Success:      true,
				Message:      "Stored metadata in Vault",
				Duration:     1200 * time.Millisecond,
			},
		},
		TotalDuration: 6210 * time.Millisecond,
		ProtocolName:  "default",
	}

	summary := buildRemediationSummary(results)

	// Check for success indicator
	if !strings.Contains(summary, "✓ Logged findings") {
		t.Errorf("missing success indicator: %s", summary)
	}

	// Check for failure indicator
	if !strings.Contains(summary, "✗ Failed to send webhook") {
		t.Errorf("missing failure indicator: %s", summary)
	}

	// Check for second success
	if !strings.Contains(summary, "✓ Stored metadata in Vault") {
		t.Errorf("missing second success: %s", summary)
	}
}

func TestBuildRemediationSummary_AllFailed(t *testing.T) {
	results := types.RemediationResults{
		Executed: true,
		Results: []types.RemediationResult{
			{
				StrategyType: "webhook",
				Success:      false,
				Message:      "Failed to send webhook: connection refused",
				Duration:     100 * time.Millisecond,
			},
			{
				StrategyType: "vault",
				Success:      false,
				Message:      "Failed to store in Vault: authentication failed",
				Duration:     50 * time.Millisecond,
			},
		},
		TotalDuration: 150 * time.Millisecond,
		ProtocolName:  "alert-critical",
	}

	summary := buildRemediationSummary(results)

	// Check that both failures are shown
	if !strings.Contains(summary, "✗ Failed to send webhook") {
		t.Errorf("missing first failure: %s", summary)
	}

	if !strings.Contains(summary, "✗ Failed to store in Vault") {
		t.Errorf("missing second failure: %s", summary)
	}

	// Should not contain any success indicators
	if strings.Contains(summary, "✓") {
		t.Errorf("unexpected success indicator in all-failed summary: %s", summary)
	}
}

func TestBuildRemediationSummary_SingleStrategy(t *testing.T) {
	results := types.RemediationResults{
		Executed: true,
		Results: []types.RemediationResult{
			{
				StrategyType: "log",
				Success:      true,
				Message:      "Logged findings",
				Duration:     5 * time.Millisecond,
			},
		},
		TotalDuration: 5 * time.Millisecond,
		ProtocolName:  "default",
	}

	summary := buildRemediationSummary(results)

	// Check singular "strategy" not "strategies"
	if !strings.Contains(summary, "1 strategy") {
		t.Errorf("expected singular 'strategy', got: %s", summary)
	}

	if strings.Contains(summary, "strategies") {
		t.Errorf("unexpected plural 'strategies' for single strategy: %s", summary)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "less than 1ms",
			duration: 500 * time.Microsecond,
			expected: "0ms",
		},
		{
			name:     "exact 1ms",
			duration: 1 * time.Millisecond,
			expected: "1ms",
		},
		{
			name:     "10ms",
			duration: 10 * time.Millisecond,
			expected: "10ms",
		},
		{
			name:     "999ms",
			duration: 999 * time.Millisecond,
			expected: "999ms",
		},
		{
			name:     "exactly 1 second",
			duration: 1000 * time.Millisecond,
			expected: "1.0s",
		},
		{
			name:     "1.5 seconds",
			duration: 1500 * time.Millisecond,
			expected: "1.5s",
		},
		{
			name:     "2.3 seconds",
			duration: 2345 * time.Millisecond,
			expected: "2.3s",
		},
		{
			name:     "5 seconds",
			duration: 5 * time.Second,
			expected: "5.0s",
		},
		{
			name:     "10.8 seconds",
			duration: 10847 * time.Millisecond,
			expected: "10.8s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}
