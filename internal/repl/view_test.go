// internal/repl/view_test.go
package repl

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpBarSpacing(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"

	view := model.View()

	lines := strings.Split(view, "\n")

	var inputLineIndex int = -1
	var helpBarLineIndex int = -1

	for i, line := range lines {
		if strings.Contains(line, "cc-start [test]>") && inputLineIndex == -1 {
			inputLineIndex = i
		}
		if strings.Contains(line, "/ commands") && helpBarLineIndex == -1 {
			helpBarLineIndex = i
		}
	}

	if inputLineIndex == -1 {
		t.Error("未找到输入提示符")
	}
	if helpBarLineIndex == -1 {
		t.Error("未找到帮助栏")
	}

	emptyLinesBetween := 0
	for i := inputLineIndex + 1; i < helpBarLineIndex; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			emptyLinesBetween++
		}
	}

	if emptyLinesBetween < 1 {
		t.Errorf("输入区和帮助栏之间应该至少有一个空行，实际有 %d 个", emptyLinesBetween)
	}
}

func TestInputPromptSpacing(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"

	view := model.View()

	if !strings.Contains(view, "cc-start [test]> ") {
		t.Error("提示符 'cc-start [test]>' 后应该有一个空格，避免与 placeholder 粘连")
	}

	if strings.Contains(view, "cc-start [test]>输入") {
		t.Error("提示符和 placeholder 之间应该有空格分隔，不应出现 'cc-start [test]>输入'")
	}

	if strings.Contains(view, "输 入") {
		t.Error("发现乱码：placeholder 中的中文字符不应被空格分隔")
	}
}

func TestViewRenderOrder(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"

	model.output.Write("这是输出内容\n多行输出")

	view := model.View()

	lines := strings.Split(view, "\n")

	var outputLineIndex int = -1
	var inputLineIndex int = -1

	for i, line := range lines {
		if strings.Contains(line, "这是输出内容") {
			outputLineIndex = i
		}
		if strings.Contains(line, "cc-start [test]>") {
			inputLineIndex = i
		}
	}

	if outputLineIndex == -1 {
		t.Error("未找到输出内容")
	}
	if inputLineIndex == -1 {
		t.Error("未找到输入提示符")
	}

	if outputLineIndex >= inputLineIndex {
		t.Errorf("输出区应该在输入区之前: 输出行 %d, 输入行 %d", outputLineIndex, inputLineIndex)
	}
}

func TestViewRenderAfterCommand(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd != nil {
		_ = cmd()
	}

	view1 := model.View()

	model.output.Write("命令输出")

	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		_ = cmd()
	}

	view2 := model.View()

	if view2 == view1 {
		t.Error("执行命令后视图应该更新")
	}

	lines := strings.Split(view2, "\n")

	var outputLineIndex int = -1
	var inputLineIndex int = -1

	for i, line := range lines {
		if strings.Contains(line, "命令输出") {
			outputLineIndex = i
		}
		if strings.Contains(line, "cc-start [test]>") {
			inputLineIndex = i
		}
	}

	if outputLineIndex == -1 {
		t.Error("未找到命令输出")
	}
	if inputLineIndex == -1 {
		t.Error("未找到输入提示符")
	}

	if outputLineIndex >= inputLineIndex {
		t.Errorf("命令执行后，输出应该在输入之前: 输出行 %d, 输入行 %d", outputLineIndex, inputLineIndex)
	}
}

// TestAutocompleteInView 测试自动补全在视图中的渲染
func TestAutocompleteInView(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"
	model.width = 80

	// 初始状态不应该有自动补全
	view := model.View()
	if strings.Contains(view, "● /") {
		t.Error("初始状态不应该显示自动补全")
	}

	// 触发自动补全
	model.input.SetValue("/")
	if model.autocomplete == nil {
		model.autocomplete = NewAutocomplete(model.Styles, model.I18n)
	}
	model.autocomplete.Show("/")

	view = model.View()
	if !strings.Contains(view, "/list") && !strings.Contains(view, "/use") {
		t.Error("显示自动补全时应该包含命令")
	}
}

// TestAutocompleteHelpBarDynamic 测试自动补全时帮助栏的动态提示
func TestAutocompleteHelpBarDynamic(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"
	model.width = 80

	// 初始帮助栏
	helpBar := model.renderHelpBar()
	if !strings.Contains(helpBar, "/ commands") {
		t.Error("初始帮助栏应该包含 '/ commands'")
	}

	// 显示自动补全时的帮助栏
	if model.autocomplete == nil {
		model.autocomplete = NewAutocomplete(model.Styles, model.I18n)
	}
	model.autocomplete.Show("/")

	helpBar = model.renderHelpBar()
	if !strings.Contains(helpBar, "tab complete") && !strings.Contains(helpBar, "tab 补全") {
		t.Errorf("自动补全显示时帮助栏应该包含 'tab complete' 或 'tab 补全'，实际是: %s", helpBar)
	}
}

// TestMultipleCommandOutputSpacing 测试多个命令输出之间应该有空行分隔
func TestMultipleCommandOutputSpacing(t *testing.T) {
	model, err := NewModel("")
	if err != nil {
		t.Fatalf("创建模型失败: %v", err)
	}

	model.currentProfile = "test"
	model.width = 80
	model.height = 24

	// 模拟执行第一个命令
	model.output.WriteInfo("$ /list")
	model.output.Write("配置1\n配置2")

	// 模拟执行第二个命令
	model.output.WriteInfo("$ /current")
	model.output.Write("当前配置: test")

	view := model.View()
	lines := strings.Split(view, "\n")

	// 查找两个命令行的位置
	var firstCmdIndex int = -1
	var secondCmdIndex int = -1

	for i, line := range lines {
		if strings.Contains(line, "$ /list") && firstCmdIndex == -1 {
			firstCmdIndex = i
		}
		if strings.Contains(line, "$ /current") && secondCmdIndex == -1 {
			secondCmdIndex = i
		}
	}

	if firstCmdIndex == -1 {
		t.Error("未找到第一个命令")
	}
	if secondCmdIndex == -1 {
		t.Error("未找到第二个命令")
	}

	// 检查两个命令之间是否有空行
	if secondCmdIndex <= firstCmdIndex {
		t.Errorf("第二个命令应该在第一个命令之后: first=%d, second=%d", firstCmdIndex, secondCmdIndex)
	}

	// 检查第一个命令的输出和第二个命令之间是否有空行
	emptyLineFound := false
	for i := firstCmdIndex + 1; i < secondCmdIndex; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			emptyLineFound = true
			break
		}
	}

	if !emptyLineFound {
		t.Error("两个命令输出之间应该有空行分隔，以便提示符正确下移")
		t.Logf("视图内容:\n%s", view)
	}
}
