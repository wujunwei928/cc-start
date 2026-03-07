// internal/repl/completer_test.go
package repl

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/c-bata/go-prompt"
)

// newTestDocument 创建测试用的 Document，设置光标位置在文本末尾
func newTestDocument(text string) prompt.Document {
	doc := prompt.NewDocument()
	doc.Text = text

	// 使用 unsafe 设置私有字段 cursorPosition
	if len(text) > 0 {
		v := reflect.ValueOf(doc).Elem()
		cursorField := v.FieldByName("cursorPosition")
		if cursorField.IsValid() {
			cursorPtr := unsafe.Pointer(cursorField.UnsafeAddr())
			*(*int)(cursorPtr) = len(text)
		}
	}

	return *doc
}

func TestCompleterCommandCompletion(t *testing.T) {
	c := NewCompleter(func() []string { return []string{"moonshot", "deepseek"} })

	// 测试命令补全
	doc := newTestDocument("li")

	suggestions := c.Complete(doc)
	if len(suggestions) == 0 {
		t.Error("expected suggestions for 'li'")
	}

	found := false
	for _, s := range suggestions {
		if s.Text == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'list' in suggestions")
	}
}

func TestCompleterProfileCompletion(t *testing.T) {
	profiles := []string{"moonshot", "deepseek", "anthropic"}
	c := NewCompleter(func() []string { return profiles })

	// 测试配置名补全 - 需要输入 "use " (带空格)
	doc := newTestDocument("use ")

	suggestions := c.Complete(doc)
	if len(suggestions) == 0 {
		t.Error("expected profile suggestions for 'use '")
	}

	found := false
	for _, s := range suggestions {
		if s.Text == "moonshot" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'moonshot' in suggestions")
	}
}

func TestCompleterAllCommands(t *testing.T) {
	c := NewCompleter(nil)

	// 测试空输入返回所有命令
	doc := newTestDocument("")
	suggestions := c.Complete(doc)

	if len(suggestions) < 10 {
		t.Errorf("expected at least 10 command suggestions, got %d", len(suggestions))
	}

	// 检查关键命令存在
	expectedCmds := []string{"list", "use", "exit", "help", "run"}
	for _, expected := range expectedCmds {
		found := false
		for _, s := range suggestions {
			if s.Text == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command '%s' in suggestions", expected)
		}
	}
}

func TestCompleterAliasCompletion(t *testing.T) {
	c := NewCompleter(nil)

	// 测试别名补全
	doc := newTestDocument("q")

	suggestions := c.Complete(doc)

	found := false
	for _, s := range suggestions {
		if s.Text == "q" { // exit 的别名
			found = true
			break
		}
	}
	if !found {
		t.Error("expected alias 'q' in suggestions")
	}
}

func TestCompleterNoProfileGetter(t *testing.T) {
	c := NewCompleter(nil) // 无 profile 获取器

	// 尝试获取 profile 补全
	doc := newTestDocument("use ")

	suggestions := c.Complete(doc)
	if len(suggestions) != 0 {
		t.Errorf("expected no suggestions without profile getter, got %d", len(suggestions))
	}
}

func TestCompleterNeedsProfileCommands(t *testing.T) {
	profiles := []string{"test-profile"}
	c := NewCompleter(func() []string { return profiles })

	// 测试所有需要 profile 的命令
	cmds := []string{"use", "switch", "show", "edit", "delete", "rm",
		"test", "default", "copy", "cp", "rename", "mv", "run"}

	for _, cmd := range cmds {
		doc := newTestDocument(cmd + " ")

		suggestions := c.Complete(doc)
		if len(suggestions) == 0 {
			t.Errorf("expected profile suggestions for '%s '", cmd)
		}

		found := false
		for _, s := range suggestions {
			if s.Text == "test-profile" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected 'test-profile' in suggestions for '%s '", cmd)
		}
	}
}

func TestCompleterUnknownCommand(t *testing.T) {
	c := NewCompleter(func() []string { return []string{"moonshot"} })

	// 测试未知命令后无补全
	doc := newTestDocument("unknown ")

	suggestions := c.Complete(doc)
	if len(suggestions) != 0 {
		t.Errorf("expected no suggestions for unknown command, got %d", len(suggestions))
	}
}

func TestCommandDefs(t *testing.T) {
	c := NewCompleter(nil)
	cmds := c.getCommandDefs()

	// 验证所有命令都有名称和描述
	for _, cmd := range cmds {
		if cmd.Name == "" {
			t.Error("command should have a name")
		}
		if cmd.Description == "" {
			t.Errorf("command '%s' should have a description", cmd.Name)
		}
	}
}

func TestCompleterCommandWithPartialMatch(t *testing.T) {
	c := NewCompleter(nil)

	// 测试部分匹配的命令
	testCases := []struct {
		input    string
		expected string
	}{
		{"ex", "exit"},
		{"he", "help"},
		{"hi", "history"},
		{"cl", "clear"},
		{"de", "default"},
	}

	for _, tc := range testCases {
		doc := newTestDocument(tc.input)
		suggestions := c.Complete(doc)

		found := false
		for _, s := range suggestions {
			if s.Text == tc.expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected '%s' in suggestions for input '%s'", tc.expected, tc.input)
		}
	}
}

// 测试 completeCommand 方法
func TestCompleteCommand(t *testing.T) {
	c := NewCompleter(nil)

	// 测试空输入返回所有命令
	suggestions := c.completeCommand(nil)
	if len(suggestions) < 10 {
		t.Errorf("expected at least 10 suggestions, got %d", len(suggestions))
	}

	// 验证包含命令和别名
	hasCmd := false
	hasAlias := false
	for _, s := range suggestions {
		if s.Text == "exit" {
			hasCmd = true
		}
		if s.Text == "q" || s.Text == "quit" {
			hasAlias = true
		}
	}
	if !hasCmd {
		t.Error("expected 'exit' command in suggestions")
	}
	if !hasAlias {
		t.Error("expected alias in suggestions")
	}
}

// 测试 completeProfile 方法
func TestCompleteProfile(t *testing.T) {
	profiles := []string{"moonshot", "deepseek"}
	c := NewCompleter(func() []string { return profiles })

	doc := newTestDocument("use ")
	suggestions := c.completeProfile([]string{"use"}, doc)

	if len(suggestions) != 2 {
		t.Errorf("expected 2 profile suggestions, got %d", len(suggestions))
	}

	// 验证 profile 名称
	found := make(map[string]bool)
	for _, s := range suggestions {
		found[s.Text] = true
	}
	if !found["moonshot"] {
		t.Error("expected 'moonshot' in suggestions")
	}
	if !found["deepseek"] {
		t.Error("expected 'deepseek' in suggestions")
	}
}
