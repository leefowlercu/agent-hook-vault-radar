package gemini

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/framework"
	"github.com/leefowlercu/agent-hook-vault-radar/pkg/types"
)

const frameworkName = "gemini"

// Framework implements the HookFramework interface for Gemini CLI
type Framework struct {
	handlers []framework.HookHandler
}

// Force compile-time check for interface implementation
var _ framework.HookFramework = (*Framework)(nil)

// NewFramework creates a new Gemini framework instance
func NewFramework() *Framework {
	f := &Framework{
		handlers: []framework.HookHandler{},
	}

	// Register default handlers
	f.RegisterHandler(NewBeforeAgentHandler())

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

// ParseInput reads and parses Gemini hook data from stdin
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

// FormatOutput formats a decision as JSON for Gemini CLI
func (f *Framework) FormatOutput(decision types.Decision, input types.HookInput) ([]byte, error) {
	output := HookOutput{}

	// Set decision field based on block status
	if decision.Block {
		output.Decision = "deny"
	} else {
		output.Decision = "allow"
	}

	// Add hook-specific output
	if hookEventName, ok := input.RawData["hook_event_name"].(string); ok {
		output.HookSpecificOutput = HookSpecificOutput{
			HookEventName: hookEventName,
		}

		// Add detailed findings to additionalContext if we're blocking
		if decision.Block {
			output.HookSpecificOutput.AdditionalContext = f.formatAdditionalContext(decision)
		}
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output; %w", err)
	}

	return data, nil
}

// formatAdditionalContext formats findings into a detailed context string
func (f *Framework) formatAdditionalContext(decision types.Decision) string {
	// Try to extract findings from metadata
	findingsAny, ok := decision.Metadata["findings"]
	if !ok {
		// No findings in metadata, use the reason message directly
		if decision.Reason != "" {
			return decision.Reason
		}
		return ""
	}

	// Type assert to []types.Finding
	findings, ok := findingsAny.([]types.Finding)
	if !ok || len(findings) == 0 {
		if decision.Reason != "" {
			return decision.Reason
		}
		return ""
	}

	// Build formatted findings message
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

// GetExitCode returns the appropriate exit code for Gemini CLI
// Always returns 0 - blocking is controlled by JSON "decision" field
func (f *Framework) GetExitCode(decision types.Decision) int {
	return 0
}

// GetName returns the framework name
func (f *Framework) GetName() string {
	return frameworkName
}
