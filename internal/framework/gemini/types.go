package gemini

// BeforeAgentInput represents the input structure for BeforeAgent hook
type BeforeAgentInput struct {
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	HookEventName string `json:"hook_event_name"`
	Timestamp     string `json:"timestamp"`
	Prompt        string `json:"prompt"`
}

// HookOutput represents the output structure for Gemini hooks
type HookOutput struct {
	Decision           string             `json:"decision"`
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOutput contains hook-specific output fields
type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}
