// internal/tui/setup/model_test.go
package setup

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// toModel 将 tea.Model 转换为 Model
func toModel(m tea.Model) Model {
	switch v := m.(type) {
	case Model:
		return v
	case *Model:
		return *v
	default:
		panic("unexpected model type")
	}
}

// TestTokenInputBackspace 测试 Token 输入框 Backspace 删除功能
func TestTokenInputBackspace(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 输入 token
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("abcdef")})
	m = toModel(result)

	// 验证输入值
	if m.tokenInput.Value() != "abcdef" {
		t.Fatalf("期望 token 值为 'abcdef'，实际为 '%s'", m.tokenInput.Value())
	}

	// 按 Backspace 删除字符
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = toModel(result)

	// 验证删除成功
	if m.tokenInput.Value() != "abcde" {
		t.Errorf("Backspace 应该删除最后一个字符，期望 'abcde'，实际 '%s'", m.tokenInput.Value())
	}
}

// TestModelInputBackspace 测试模型输入框 Backspace 删除功能
func TestModelInputBackspace(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 验证预设的模型值已设置
	initialModel := m.modelInput.Value()
	if initialModel == "" {
		t.Fatal("预设模型值不应为空")
	}

	// 输入 token 并进入模型步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputModel {
		t.Fatalf("期望步骤为 stepInputModel，实际为 %d", m.step)
	}

	// 按 Backspace 删除字符
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = toModel(result)

	// 验证删除成功
	if m.modelInput.Value() == initialModel {
		t.Errorf("Backspace 应该删除字符，但值未改变: %s", m.modelInput.Value())
	}
}

// TestNameInputBackspace 测试自定义模式下名称输入框 Backspace 删除功能
func TestNameInputBackspace(t *testing.T) {
	m := InitialModel()

	// 选择 "自定义" 选项（最后一项）
	m.selected = len(m.presets) - 1
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputName {
		t.Fatalf("期望步骤为 stepInputName，实际为 %d", m.step)
	}

	// 输入名称
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("myconfig")})
	m = toModel(result)

	if m.nameInput.Value() != "myconfig" {
		t.Fatalf("期望名称为 'myconfig'，实际为 '%s'", m.nameInput.Value())
	}

	// 按 Backspace 删除字符
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = toModel(result)

	// 验证删除成功
	if m.nameInput.Value() != "myconfi" {
		t.Errorf("Backspace 应该删除最后一个字符，期望 'myconfi'，实际 '%s'", m.nameInput.Value())
	}
}

// TestModelInputCanType 测试模型输入框可以输入新字符
func TestModelInputCanType(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 输入 token 并进入下一步
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputModel {
		t.Fatalf("期望步骤为 stepInputModel，实际为 %d", m.step)
	}

	// 尝试输入新字符
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-new")})
	m = toModel(result)

	// 验证新字符已添加
	value := m.modelInput.Value()
	if len(value) < len("claude-sonnet-4-5-20250929") {
		t.Errorf("输入字符应该被添加，但值变短了: %s", value)
	}
}

// TestEscGoBackFromModel 测试 ESC 从模型步骤返回上一步
func TestEscGoBackFromModel(t *testing.T) {
	m := InitialModel()

	// 选择预设，输入 token，到达模型步骤
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputModel {
		t.Fatalf("期望步骤为 stepInputModel，实际为 %d", m.step)
	}

	// 按 ESC 返回上一步
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = toModel(result)

	if m.step != stepInputToken {
		t.Errorf("ESC 应该返回 token 步骤，期望 stepInputToken，实际 %d", m.step)
	}
}

// TestEscGoBackFromToken 测试 ESC 从 Token 步骤返回上一步
func TestEscGoBackFromToken(t *testing.T) {
	m := InitialModel()

	// 选择预设，到达 token 步骤
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputToken {
		t.Fatalf("期望步骤为 stepInputToken，实际为 %d", m.step)
	}

	// 按 ESC 返回预设选择步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = toModel(result)

	if m.step != stepSelectPreset {
		t.Errorf("ESC 应该返回预设选择步骤，期望 stepSelectPreset，实际 %d", m.step)
	}
}

// TestEscGoBackFromName 测试 ESC 从名称步骤返回上一步（自定义模式）
func TestEscGoBackFromName(t *testing.T) {
	m := InitialModel()

	// 选择 "自定义"
	m.selected = len(m.presets) - 1
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputName {
		t.Fatalf("期望步骤为 stepInputName，实际为 %d", m.step)
	}

	// 按 ESC 返回预设选择步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = toModel(result)

	if m.step != stepSelectPreset {
		t.Errorf("ESC 应该返回预设选择步骤，期望 stepSelectPreset，实际 %d", m.step)
	}
}
