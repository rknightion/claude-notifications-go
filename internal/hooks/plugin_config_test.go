package hooks

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

type pluginHookSettings struct {
	Hooks map[string][]pluginHookMatcherGroup `json:"hooks"`
}

type pluginHookMatcherGroup struct {
	Matcher string              `json:"matcher,omitempty"`
	Hooks   []pluginHookCommand `json:"hooks"`
}

type pluginHookCommand struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Timeout int      `json:"timeout"`
	Shell   string   `json:"shell,omitempty"`
}

func TestPluginHooksUseExecFormWrapper(t *testing.T) {
	data, err := os.ReadFile("../../hooks/claude-hooks.json")
	if err != nil {
		t.Fatal(err)
	}

	raw := string(data)
	for _, forbidden := range []string{
		"hook-wrapper.sh handle-hook",
		"$input",
		"powershell",
		`"shell"`,
	} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("default hooks config contains %q", forbidden)
		}
	}

	var settings pluginHookSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"PreToolUse":   "PreToolUse",
		"Notification": "Notification",
		"Stop":         "Stop",
		"SubagentStop": "SubagentStop",
		"TeammateIdle": "TeammateIdle",
	}

	for hookEvent, expectedArg := range expected {
		groups := settings.Hooks[hookEvent]
		if len(groups) != 1 {
			t.Fatalf("%s groups = %d, want 1", hookEvent, len(groups))
		}
		if len(groups[0].Hooks) != 1 {
			t.Fatalf("%s commands = %d, want 1", hookEvent, len(groups[0].Hooks))
		}

		hook := groups[0].Hooks[0]
		if hook.Type != "command" {
			t.Fatalf("%s type = %q, want command", hookEvent, hook.Type)
		}
		if hook.Command != "sh" {
			t.Fatalf("%s command = %q, want sh", hookEvent, hook.Command)
		}
		wantArgs := []string{"${CLAUDE_PLUGIN_ROOT}/bin/hook-wrapper.sh", "handle-hook", expectedArg}
		if !reflect.DeepEqual(hook.Args, wantArgs) {
			t.Fatalf("%s args = %#v, want %#v", hookEvent, hook.Args, wantArgs)
		}
		if hook.Shell != "" {
			t.Fatalf("%s shell = %q, want empty", hookEvent, hook.Shell)
		}
		if hook.Timeout != 30 {
			t.Fatalf("%s timeout = %d, want 30", hookEvent, hook.Timeout)
		}
	}
}

func TestCodexPluginHooksUseOnlySupportedEvents(t *testing.T) {
	data, err := os.ReadFile("../../hooks/hooks.json")
	if err != nil {
		t.Fatal(err)
	}

	var settings pluginHookSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}

	supported := map[string]bool{
		"UserPromptSubmit":  true,
		"PermissionRequest": true,
		"PostToolUse":       true,
		"Stop":              true,
	}
	for event := range settings.Hooks {
		if !supported[event] {
			t.Fatalf("Codex hook registry contains unsupported event %q", event)
		}
	}
	for event := range supported {
		if len(settings.Hooks[event]) == 0 {
			t.Errorf("Codex hook registry is missing %q", event)
		}
	}

	raw := string(data)
	if !strings.Contains(raw, "${PLUGIN_ROOT}/bin/hook-wrapper.sh") {
		t.Fatal("Codex hooks must resolve the wrapper through PLUGIN_ROOT")
	}
	if !strings.Contains(raw, "--runtime codex") {
		t.Fatal("Codex hooks must pass the runtime boundary explicitly")
	}
	for _, forbidden := range []string{"Notification", "TeammateIdle", "${CLAUDE_PLUGIN_ROOT}"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("Codex hook registry contains %q", forbidden)
		}
	}
}

func TestRuntimePluginManifestsKeepHookRegistriesSeparate(t *testing.T) {
	claudeData, err := os.ReadFile("../../.claude-plugin/plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	var claudeManifest map[string]any
	if err := json.Unmarshal(claudeData, &claudeManifest); err != nil {
		t.Fatal(err)
	}
	if got := claudeManifest["hooks"]; got != "./hooks/claude-hooks.json" {
		t.Fatalf("Claude hooks = %#v, want separate Claude registry", got)
	}

	codexData, err := os.ReadFile("../../.codex-plugin/plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	var codexManifest map[string]any
	if err := json.Unmarshal(codexData, &codexManifest); err != nil {
		t.Fatal(err)
	}
	if _, present := codexManifest["hooks"]; present {
		t.Fatal("Codex manifest must use the supported default hooks/hooks.json discovery path")
	}
}
