// internal/repl/history_test.go
package repl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryAdd(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
		commands: make([]string, 0),
	}

	h.Add("list")
	h.Add("use moonshot")

	commands := h.GetCommands()
	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}
	if commands[0] != "list" {
		t.Errorf("expected first command 'list', got '%s'", commands[0])
	}
}

func TestHistoryNoDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
		commands: make([]string, 0),
	}

	h.Add("list")
	h.Add("list") // 重复命令

	commands := h.GetCommands()
	if len(commands) != 1 {
		t.Errorf("expected 1 command (no duplicate), got %d", len(commands))
	}
}

func TestHistoryMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
		commands: make([]string, 0),
	}

	// 添加超过最大限制的命令
	for i := 0; i < maxHistorySize+100; i++ {
		h.Add("command")
	}

	commands := h.GetCommands()
	if len(commands) > maxHistorySize {
		t.Errorf("expected max %d commands, got %d", maxHistorySize, len(commands))
	}
}

func TestHistoryPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "history")

	// 创建并添加命令
	h1 := &History{
		filePath: filePath,
		commands: make([]string, 0),
	}
	h1.Add("list")
	h1.Add("use moonshot")

	// 创建新的历史实例，应该加载之前的记录
	h2 := &History{
		filePath: filePath,
		commands: make([]string, 0),
	}
	h2.load()

	commands := h2.GetCommands()
	if len(commands) != 2 {
		t.Errorf("expected 2 commands after reload, got %d", len(commands))
	}
}

func TestHistoryEmptyCommand(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
		commands: make([]string, 0),
	}

	h.Add("")   // 空命令
	h.Add("  ") // 仅空格

	commands := h.GetCommands()
	if len(commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(commands))
	}
}

func TestHistoryTrimSpace(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
		commands: make([]string, 0),
	}

	h.Add("  list  ")

	commands := h.GetCommands()
	if len(commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(commands))
	}
	if commands[0] != "list" {
		t.Errorf("expected trimmed command 'list', got '%s'", commands[0])
	}
}

func TestHistoryLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "nonexistent", "history"),
		commands: make([]string, 0),
	}

	// 加载不存在的文件不应报错
	h.load()

	commands := h.GetCommands()
	if len(commands) != 0 {
		t.Errorf("expected 0 commands from nonexistent file, got %d", len(commands))
	}
}

func TestNewHistory(t *testing.T) {
	h := NewHistory()
	if h == nil {
		t.Fatal("expected non-nil History")
	}

	// 验证文件路径包含用户目录
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".cc-start", historyFile)
	if h.filePath != expectedPath {
		t.Errorf("expected filePath '%s', got '%s'", expectedPath, h.filePath)
	}
}
