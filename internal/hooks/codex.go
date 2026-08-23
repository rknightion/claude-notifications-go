package hooks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/777genius/claude-notifications/internal/analyzer"
)

// handleCodexHook adapts documented Codex hook fields without reading the
// transcript_path, whose format is explicitly not a stable hook interface.
func (h *Handler) handleCodexHook(hookEvent string, hookData *HookData) (analyzer.Status, error) {
	switch hookEvent {
	case "UserPromptSubmit":
		if err := h.stateMgr.UpdatePrompt(hookData.SessionID, hookData.Prompt, hookData.CWD); err != nil {
			return analyzer.StatusUnknown, fmt.Errorf("record Codex prompt state: %w", err)
		}
		return analyzer.StatusUnknown, nil
	case "PermissionRequest":
		return analyzer.StatusQuestion, nil
	case "PostToolUse":
		return codexAPIErrorStatus(hookData.ToolResponse), nil
	case "Stop":
		message := strings.TrimSpace(hookData.LastAssistantMessage)
		if codexSessionLimitMessage(message) {
			return analyzer.StatusSessionLimitReached, nil
		}
		if codexAPIErrorMessage(message) {
			if status := codexAPIErrorTextStatus(message); status != analyzer.StatusUnknown {
				return status, nil
			}
		}
		if hookData.PermissionMode == "plan" {
			return analyzer.StatusPlanReady, nil
		}
		stateData, err := h.stateMgr.Load(hookData.SessionID)
		if err != nil {
			return analyzer.StatusUnknown, fmt.Errorf("load Codex prompt state: %w", err)
		}
		if stateData != nil && codexReviewPrompt(stateData.LastPrompt) && message != "" {
			return analyzer.StatusReviewComplete, nil
		}
		if message != "" {
			return analyzer.StatusTaskComplete, nil
		}
		return analyzer.StatusUnknown, nil
	default:
		return analyzer.StatusUnknown, fmt.Errorf("unknown Codex hook event: %s", hookEvent)
	}
}

func codexReviewPrompt(prompt string) bool {
	prompt = strings.ToLower(prompt)
	return strings.Contains(prompt, "review") || strings.Contains(prompt, "audit")
}

func codexSessionLimitMessage(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "session limit reached") ||
		strings.Contains(message, "session limit has been reached") ||
		strings.Contains(message, "usage limit reached")
}

func codexAPIErrorStatus(response json.RawMessage) analyzer.Status {
	if len(response) == 0 {
		return analyzer.StatusUnknown
	}
	text := strings.ToLower(string(response))
	if !strings.Contains(text, `"iserror":true`) &&
		!strings.Contains(text, `"is_error":true`) &&
		!strings.Contains(text, "api error") {
		return analyzer.StatusUnknown
	}
	return codexAPIErrorTextStatus(text)
}

func codexAPIErrorMessage(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "api error") ||
		strings.Contains(message, "api request failed")
}

func codexAPIErrorTextStatus(text string) analyzer.Status {
	text = strings.ToLower(text)
	if strings.Contains(text, "rate limit") || strings.Contains(text, "overloaded") ||
		strings.Contains(text, "429") || strings.Contains(text, "529") {
		return analyzer.StatusAPIErrorOverloaded
	}
	if strings.Contains(text, "authentication") || strings.Contains(text, "unauthorized") ||
		strings.Contains(text, "401") || strings.Contains(text, "api error") {
		return analyzer.StatusAPIError
	}
	return analyzer.StatusUnknown
}
