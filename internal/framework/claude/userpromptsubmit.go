package claude

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/framework"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

const userPromptSubmitType = "UserPromptSubmit"

// UserPromptSubmitHandler handles the UserPromptSubmit hook
type UserPromptSubmitHandler struct{}

// Force compile-time check for interface implementation
var _ framework.HookHandler = (*UserPromptSubmitHandler)(nil)

// NewUserPromptSubmitHandler creates a new UserPromptSubmit handler
func NewUserPromptSubmitHandler() *UserPromptSubmitHandler {
	return &UserPromptSubmitHandler{}
}

// ExtractContent extracts the prompt text for scanning
func (h *UserPromptSubmitHandler) ExtractContent(ctx context.Context, input types.HookInput) (types.ScanContent, error) {
	var promptInput UserPromptSubmitInput

	// Marshal and unmarshal to convert map to struct
	data, err := json.Marshal(input.RawData)
	if err != nil {
		return types.ScanContent{}, fmt.Errorf("failed to marshal input data; %w", err)
	}

	if err := json.Unmarshal(data, &promptInput); err != nil {
		return types.ScanContent{}, fmt.Errorf("failed to unmarshal UserPromptSubmit input; %w", err)
	}

	return types.ScanContent{
		Type:    "text",
		Content: promptInput.Prompt,
		Metadata: map[string]string{
			"session_id":      promptInput.SessionID,
			"transcript_path": promptInput.TranscriptPath,
			"cwd":             promptInput.CWD,
		},
	}, nil
}

// MakeDecision creates a decision based on scan results
func (h *UserPromptSubmitHandler) MakeDecision(ctx context.Context, results types.ScanResults, input types.HookInput) (types.Decision, error) {
	decision := types.Decision{
		Block:    false,
		ExitCode: 0,
		Metadata: make(map[string]any),
	}

	if results.Error != nil {
		// If scanning failed, we'll allow by default but log the error
		decision.Metadata["scan_error"] = results.Error.Error()
		return decision, nil
	}

	if results.HasFindings {
		decision.Block = true
		decision.ExitCode = 2

		// Build a detailed reason message
		reason := fmt.Sprintf("Vault Radar detected %d security finding(s) in your prompt:\n\n", len(results.Findings))
		for i, finding := range results.Findings {
			reason += fmt.Sprintf("%d. [%s] %s: %s\n", i+1, finding.Severity, finding.Type, finding.Description)
		}
		reason += "\nPlease remove or redact sensitive information before submitting."

		decision.Reason = reason
		decision.Metadata["findings"] = results.Findings
		decision.Metadata["finding_count"] = len(results.Findings)
	}

	return decision, nil
}

// GetType returns the hook type name
func (h *UserPromptSubmitHandler) GetType() string {
	return userPromptSubmitType
}

// CanHandle returns true if this handler can process the given hook input
func (h *UserPromptSubmitHandler) CanHandle(input types.HookInput) bool {
	return input.Framework == "claude" && input.HookType == userPromptSubmitType
}
