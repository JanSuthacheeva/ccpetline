package pet

import "encoding/json"

// ClaudeInput is the JSON payload Claude Code writes to the statusline's and
// the hook's stdin. Both binaries receive the same shape; fields that only
// apply to one context (e.g. tool_name for hooks) are zero elsewhere.
type ClaudeInput struct {
	SessionID     string `json:"session_id"`
	HookEventName string `json:"hook_event_name"`
	Source        string `json:"source"`
	ToolName      string `json:"tool_name"`
	Model         struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"model"`
	ContextWindow struct {
		// Pointer so an absent field is distinguishable from 0%.
		UsedPercentage *float64 `json:"used_percentage"`
	} `json:"context_window"`
	Cost struct {
		TotalCostUSD float64 `json:"total_cost_usd"`
	} `json:"cost"`
	RateLimits map[string]any `json:"rate_limits"`
}

// ParseClaudeInput decodes a stdin payload. Returns nil when the data is not
// valid JSON, which callers treat as "no payload".
func ParseClaudeInput(data []byte) *ClaudeInput {
	var in ClaudeInput
	if err := json.Unmarshal(data, &in); err != nil {
		return nil
	}
	return &in
}
