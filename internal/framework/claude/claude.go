package claude

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/framework"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

const frameworkName = "claude"

// Framework implements the HookFramework interface for Claude Code
type Framework struct {
	handlers []framework.HookHandler
}

// Force compile-time check for interface implementation
var _ framework.HookFramework = (*Framework)(nil)

// NewFramework creates a new Claude framework instance
func NewFramework() *Framework {
	f := &Framework{
		handlers: []framework.HookHandler{},
	}

	// Register default handlers
	f.RegisterHandler(NewUserPromptSubmitHandler())

	return f
}

// RegisterHandler registers a hook handler with the framework
func (f *Framework) RegisterHandler(handler framework.HookHandler) {
	f.handlers = append(f.handlers, handler)
}

// GetHandler returns the appropriate handler for the given input
func (f *Framework) GetHandler(input types.HookInput) (framework.HookHandler, error) {
	for _, handler := range f.handlers {
		if handler.CanHandle(input) {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("no handler found for hook type %q", input.HookType)
}

// ParseInput reads and parses Claude hook data from stdin
func (f *Framework) ParseInput(reader io.Reader) (types.HookInput, error) {
	var rawData map[string]any

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&rawData); err != nil {
		return types.HookInput{}, fmt.Errorf("failed to decode JSON input; %w", err)
	}

	// Extract hook event name
	hookEventName, ok := rawData["hook_event_name"].(string)
	if !ok {
		return types.HookInput{}, fmt.Errorf("missing or invalid hook_event_name")
	}

	return types.HookInput{
		Framework: frameworkName,
		HookType:  hookEventName,
		RawData:   rawData,
	}, nil
}

// FormatOutput formats a decision as JSON for Claude Code
func (f *Framework) FormatOutput(decision types.Decision, input types.HookInput) ([]byte, error) {
	output := HookOutput{
		Continue:       !decision.Block,
		SuppressOutput: false,
	}

	if decision.Block {
		output.Decision = "block"
		output.Reason = decision.Reason
		output.SystemMessage = decision.Reason
	}

	// Add hook-specific output if available
	if hookEventName, ok := input.RawData["hook_event_name"].(string); ok {
		output.HookSpecificOutput = HookSpecificOutput{
			HookEventName: hookEventName,
		}
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output; %w", err)
	}

	return data, nil
}

// GetName returns the framework name
func (f *Framework) GetName() string {
	return frameworkName
}
