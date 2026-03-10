// cmd/root_test.go
package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
	}

	if rootCmd.Use != "cc-start" {
		t.Errorf("expected Use 'cc-start', got '%s'", rootCmd.Use)
	}
}

func TestLaunchCommandExists(t *testing.T) {
	launchCmdFound := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "launch" {
			launchCmdFound = true
			break
		}
	}
	if !launchCmdFound {
		t.Error("launch command should exist")
	}
}

func TestFindDashSeparator(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{
			name:     "no separator",
			args:     []string{"cc-start", "MiniMax"},
			expected: -1,
		},
		{
			name:     "separator at end",
			args:     []string{"cc-start", "MiniMax", "--"},
			expected: 2,
		},
		{
			name:     "separator with args after",
			args:     []string{"cc-start", "MiniMax", "--", "--help"},
			expected: 2,
		},
		{
			name:     "separator with multiple args after",
			args:     []string{"cc-start", "MiniMax", "--", "--dangerously-skip-permissions", "--model", "claude-3"},
			expected: 2,
		},
		{
			name:     "separator without profile",
			args:     []string{"cc-start", "--", "--help"},
			expected: 1,
		},
		{
			name:     "separator at start",
			args:     []string{"cc-start", "--"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findDashSeparator(tt.args)
			if result != tt.expected {
				t.Errorf("findDashSeparator() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestIsFlag(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"--help", true},
		{"-h", true},
		{"--dangerously-skip-permissions", true},
		{"MiniMax", false},
		{"", false},
		{"profile-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			result := isFlag(tt.arg)
			if result != tt.expected {
				t.Errorf("isFlag(%q) = %v, expected %v", tt.arg, result, tt.expected)
			}
		})
	}
}
