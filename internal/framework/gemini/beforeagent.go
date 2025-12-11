package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/framework"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

const beforeAgentType = "BeforeAgent"

// BeforeAgentHandler handles the BeforeAgent hook
type BeforeAgentHandler struct{}

// Force compile-time check for interface implementation
var _ framework.HookHandler = (*BeforeAgentHandler)(nil)

// NewBeforeAgentHandler creates a new BeforeAgent handler
func NewBeforeAgentHandler() *BeforeAgentHandler {
	return &BeforeAgentHandler{}
}

// ExtractContent extracts the prompt text for scanning
func (h *BeforeAgentHandler) ExtractContent(ctx context.Context, input types.HookInput) (types.ScanContent, error) {
	var promptInput BeforeAgentInput

	// Marshal and unmarshal to convert map to struct
	data, err := json.Marshal(input.RawData)
	if err != nil {
		return types.ScanContent{}, fmt.Errorf("failed to marshal input data; %w", err)
	}

	if err := json.Unmarshal(data, &promptInput); err != nil {
		return types.ScanContent{}, fmt.Errorf("failed to unmarshal BeforeAgent input; %w", err)
	}

	return types.ScanContent{
		Type:    "text",
		Content: promptInput.Prompt,
		Metadata: map[string]string{
			"session_id": promptInput.SessionID,
			"cwd":        promptInput.CWD,
			"timestamp":  promptInput.Timestamp,
		},
	}, nil
}

// GetType returns the hook type name
func (h *BeforeAgentHandler) GetType() string {
	return beforeAgentType
}

// CanHandle returns true if this handler can process the given hook input
func (h *BeforeAgentHandler) CanHandle(input types.HookInput) bool {
	return input.Framework == "gemini" && input.HookType == beforeAgentType
}
