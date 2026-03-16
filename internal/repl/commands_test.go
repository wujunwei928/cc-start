// internal/repl/commands_test.go
package repl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wujunwei928/cc-start/internal/config"
)

// 测试辅助函数：创建临时配置
func setupTestREPL(t *testing.T) (*REPL, string) {
	t.Helper()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "settings.json")

	// 创建测试配置
	cfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "test1", AnthropicBaseURL: "https://api.test1.com", Token: "token1", Model: "model1"},
			{Name: "test2", AnthropicBaseURL: "https://api.test2.com", Token: "token2", Model: "model2"},
		},
		Default: "test1",
	}

	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("保存测试配置失败: %v", err)
	}

	repl, err := New(cfgPath)
	if err != nil {
		t.Fatalf("创建 REPL 失败: %v", err)
	}

	return repl, cfgPath
}

// TestCmdList 测试 list 命令
func TestCmdList(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试正常列出
	repl.cmdList(nil)

	// 测试空配置
	repl.cfg.Profiles = nil
	repl.cmdList(nil)
}

// TestCmdUse 测试 use 命令
func TestCmdUse(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试正常切换
	repl.cmdUse([]string{"test2"})
	if repl.currentName != "test2" {
		t.Errorf("期望 currentName 为 test2，实际为 %s", repl.currentName)
	}

	// 测试不存在的配置
	repl.cmdUse([]string{"nonexistent"})
	// 应该不会崩溃，currentName 保持不变

	// 测试无参数
	repl.cmdUse(nil)
}

// TestCmdCurrent 测试 current 命令
func TestCmdCurrent(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试显示当前配置
	repl.cmdCurrent(nil)

	// 测试无当前配置
	repl.currentName = ""
	repl.cmdCurrent(nil)
}

// TestCmdDefault 测试 default 命令
func TestCmdDefault(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试显示默认配置
	repl.cmdDefault(nil)

	// 测试设置默认配置
	repl.cmdDefault([]string{"test2"})
	if repl.cfg.Default != "test2" {
		t.Errorf("期望 Default 为 test2，实际为 %s", repl.cfg.Default)
	}

	// 测试设置不存在的配置
	repl.cmdDefault([]string{"nonexistent"})
}

// TestCmdShow 测试 show 命令
func TestCmdShow(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试显示指定配置
	repl.cmdShow([]string{"test1"})

	// 测试显示当前配置
	repl.currentName = "test2"
	repl.cmdShow(nil)

	// 测试无参数且无当前配置
	repl.currentName = ""
	repl.cmdShow(nil)

	// 测试不存在的配置
	repl.cmdShow([]string{"nonexistent"})
}

// TestCmdAdd 测试 add 命令
func TestCmdAdd(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// add 命令应该只显示提示信息
	repl.cmdAdd(nil)
}

// TestCmdEdit 测试 edit 命令
func TestCmdEdit(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试无参数且无当前配置
	repl.currentName = ""
	repl.cmdEdit(nil)

	// 测试指定不存在的配置
	repl.cmdEdit([]string{"nonexistent"})
}

// TestCmdDelete 测试 delete 命令
func TestCmdDelete(t *testing.T) {
	repl, _ := setupTestREPL(t)

	initialCount := len(repl.cfg.Profiles)

	// 测试无参数
	repl.cmdDelete(nil)
	if len(repl.cfg.Profiles) != initialCount {
		t.Error("无参数时不应删除配置")
	}

	// 测试不存在的配置
	repl.cmdDelete([]string{"nonexistent"})
	if len(repl.cfg.Profiles) != initialCount {
		t.Error("不存在的配置不应影响配置列表")
	}
}

// TestCmdCopy 测试 copy 命令
func TestCmdCopy(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试参数不足
	repl.cmdCopy(nil)
	repl.cmdCopy([]string{"test1"})

	// 测试复制到已存在的名称
	repl.cmdCopy([]string{"test1", "test2"})

	// 测试从不存在的配置复制
	repl.cmdCopy([]string{"nonexistent", "test3"})
}

// TestCmdRename 测试 rename 命令
func TestCmdRename(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试参数不足
	repl.cmdRename(nil)
	repl.cmdRename([]string{"test1"})

	// 测试重命名到已存在的名称
	repl.cmdRename([]string{"test1", "test2"})

	// 测试从不存在的配置重命名
	repl.cmdRename([]string{"nonexistent", "test3"})
}

// TestCmdTest 测试 test 命令
func TestCmdTest(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试指定配置（可能因网络原因失败，但不应崩溃）
	repl.cmdTest([]string{"test1"})

	// 测试不存在的配置
	repl.cmdTest([]string{"nonexistent"})

	// 测试无参数且无当前配置
	repl.currentName = ""
	repl.cmdTest(nil)
}

// TestCmdExport 测试 export 命令
func TestCmdExport(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试导出到 stdout
	repl.cmdExport(nil)

	// 测试导出到文件
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.json")
	repl.cmdExport([]string{exportPath})

	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("导出文件应该存在")
	}
}

// TestCmdImport 测试 import 命令
func TestCmdImport(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试无参数
	repl.cmdImport(nil)

	// 测试不存在的文件
	repl.cmdImport([]string{"nonexistent.json"})

	// 创建测试导入文件
	tmpDir := t.TempDir()
	importPath := filepath.Join(tmpDir, "import.json")
	importCfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "imported", AnthropicBaseURL: "https://imported.com", Token: "token"},
		},
	}
	data, err := json.Marshal(importCfg)
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}
	os.WriteFile(importPath, data, 0644)

	repl.cmdImport([]string{importPath})
}

// TestCmdHistory 测试 history 命令
func TestCmdHistory(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试空历史
	repl.cmdHistory(nil)

	// 添加一些历史
	repl.history.Add("list")
	repl.history.Add("use test1")
	repl.cmdHistory(nil)
}

// TestCmdHelp 测试 help 命令
func TestCmdHelp(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试显示所有命令
	repl.cmdHelp(nil)

	// 测试显示特定命令帮助（支持带或不带 / 前缀）
	commands := []string{"/list", "/use", "/run", "/edit", "/delete", "/copy", "/rename"}
	for _, cmd := range commands {
		repl.cmdHelp([]string{cmd})
	}

	// 测试别名
	repl.cmdHelp([]string{"/ls"})
	repl.cmdHelp([]string{"/?"})

	// 测试不带 / 前缀（兼容性）
	repl.cmdHelp([]string{"list"})
	repl.cmdHelp([]string{"use"})

	// 测试未知命令
	repl.cmdHelp([]string{"unknown"})
}

// TestCmdClear 测试 clear 命令
func TestCmdClear(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// clear 命令不应崩溃
	repl.cmdClear(nil)
}

// TestExecuteCommand 测试命令分发
func TestExecuteCommand(t *testing.T) {
	repl, _ := setupTestREPL(t)

	// 测试所有命令及其别名（使用 / 前缀）
	testCases := []struct {
		cmd  string
		args []string
	}{
		{"/list", nil},
		{"/ls", nil},
		{"/use", []string{"test1"}},
		{"/switch", []string{"test1"}},
		{"/current", nil},
		{"/status", nil},
		{"/default", nil},
		{"/show", []string{"test1"}},
		{"/add", nil},
		{"/new", nil},
		{"/edit", nil},
		{"/delete", nil}, // 不传参数，避免实际删除
		{"/rm", nil},
		{"/copy", nil},
		{"/cp", nil},
		{"/rename", nil},
		{"/mv", nil},
		{"/test", nil},
		{"/export", nil},
		{"/import", nil},
		{"/history", nil},
		{"/help", nil},
		{"/?", nil},
		{"/h", nil},
		{"/clear", nil},
		{"/cls", nil},
		{"/setup", nil},
		{"/unknown", nil}, // 未知命令
	}

	for _, tc := range testCases {
		repl.ExecuteCommand(tc.cmd, tc.args)
	}
}

// TestMaskAPIKey 测试 API Key 遮蔽
func TestMaskAPIKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"short", "****"},                    // 长度 <= 8，返回 ****
		{"12345678", "****"},                 // 长度 = 8，返回 ****
		{"123456789", "1234****6789"},        // 长度 9，遮蔽中间
		{"1234567890123456", "1234****3456"}, // 长度 16，遮蔽中间
		{"", "****"},                         // 空字符串，返回 ****
	}

	for _, tc := range testCases {
		result := maskAPIKey(tc.input)
		if result != tc.expected {
			t.Errorf("maskAPIKey(%q) = %q, 期望 %q", tc.input, result, tc.expected)
		}
	}
}
