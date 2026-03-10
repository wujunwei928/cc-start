// internal/tools/tools_test.go
package tools

import (
	"testing"
)

func TestGetTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		wantErr  bool
		wantExec string
	}{
		{
			name:     "get claude tool",
			toolName: "claude",
			wantErr:  false,
			wantExec: "claude",
		},
		{
			name:     "get codex tool",
			toolName: "codex",
			wantErr:  false,
			wantExec: "codex",
		},
		{
			name:     "get opencode tool",
			toolName: "opencode",
			wantErr:  false,
			wantExec: "opencode",
		},
		{
			name:     "unknown tool",
			toolName: "unknown",
			wantErr:  true,
			wantExec: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := GetTool(tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tool.Executable != tt.wantExec {
				t.Errorf("GetTool().Executable = %v, want %v", tool.Executable, tt.wantExec)
			}
		})
	}
}

func TestToolGetEnvName(t *testing.T) {
	tool, _ := GetTool("claude")

	tests := []struct {
		param    string
		expected string
	}{
		{ParamToken, "ANTHROPIC_AUTH_TOKEN"},
		{ParamBaseURL, "ANTHROPIC_BASE_URL"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.param, func(t *testing.T) {
			got := tool.GetEnvName(tt.param)
			if got != tt.expected {
				t.Errorf("GetEnvName(%s) = %v, want %v", tt.param, got, tt.expected)
			}
		})
	}
}

func TestListTools(t *testing.T) {
	tools := ListTools()
	if len(tools) != 3 {
		t.Errorf("ListTools() returned %d tools, want 3", len(tools))
	}
}
