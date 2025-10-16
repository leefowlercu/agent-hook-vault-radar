package claude

// UserPromptSubmitInput represents the input structure for UserPromptSubmit hook
type UserPromptSubmitInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	Prompt         string `json:"prompt"`
}

// HookOutput represents the common output structure for Claude hooks
type HookOutput struct {
	Decision           string             `json:"decision,omitempty"`
	Reason             string             `json:"reason,omitempty"`
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
	Continue           bool               `json:"continue"`
	StopReason         string             `json:"stopReason,omitempty"`
	SuppressOutput     bool               `json:"suppressOutput"`
	SystemMessage      string             `json:"systemMessage,omitempty"`
}

// HookSpecificOutput contains hook-specific output fields
type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}
