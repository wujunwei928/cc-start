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

	// 确认名称步骤（预设名已自动填充）
	if m.step != stepInputName {
		t.Fatalf("期望步骤为 stepInputName，实际为 %d", m.step)
	}
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
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

// TestSonnetInputBackspace 测试 Sonnet 模型输入框 Backspace 删除功能
func TestSonnetInputBackspace(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 确认名称步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 验证预设的 sonnet 模型值已设置
	initialSonnet := m.sonnetInput.Value()
	if initialSonnet == "" {
		t.Fatal("预设 sonnet 模型值不应为空")
	}

	// 输入 token 并经过 haiku 步骤到达 sonnet 步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputHaikuModel {
		t.Fatalf("期望步骤为 stepInputHaikuModel，实际为 %d", m.step)
	}

	// 跳过 haiku 步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputSonnetModel {
		t.Fatalf("期望步骤为 stepInputSonnetModel，实际为 %d", m.step)
	}

	// 按 Backspace 删除字符
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = toModel(result)

	// 验证删除成功
	if m.sonnetInput.Value() == initialSonnet {
		t.Errorf("Backspace 应该删除字符，但值未改变: %s", m.sonnetInput.Value())
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

// TestSonnetInputCanType 测试 Sonnet 模型输入框可以输入新字符
func TestSonnetInputCanType(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 确认名称步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 输入 token 并经过 haiku 步骤到达 sonnet 步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputSonnetModel {
		t.Fatalf("期望步骤为 stepInputSonnetModel，实际为 %d", m.step)
	}

	// 尝试输入新字符
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-new")})
	m = toModel(result)

	// 验证新字符已添加
	value := m.sonnetInput.Value()
	if len(value) < len("claude-sonnet-4-5-20250929") {
		t.Errorf("输入字符应该被添加，但值变短了: %s", value)
	}
}

// TestEscGoBackFromSonnet 测试 ESC 从 Sonnet 步骤返回 Haiku 步骤
func TestEscGoBackFromSonnet(t *testing.T) {
	m := InitialModel()

	// 选择预设，确认名称，输入 token，经过 haiku 到达 sonnet 步骤
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputSonnetModel {
		t.Fatalf("期望步骤为 stepInputSonnetModel，实际为 %d", m.step)
	}

	// 按 ESC 返回上一步（Haiku）
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = toModel(result)

	if m.step != stepInputHaikuModel {
		t.Errorf("ESC 应该返回 haiku 步骤，期望 stepInputHaikuModel，实际 %d", m.step)
	}
}

// TestEscGoBackFromToken 测试 ESC 从 Token 步骤返回上一步
func TestEscGoBackFromToken(t *testing.T) {
	m := InitialModel()

	// 选择预设，确认名称，到达 token 步骤
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputToken {
		t.Fatalf("期望步骤为 stepInputToken，实际为 %d", m.step)
	}

	// 按 ESC 返回名称步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = toModel(result)

	if m.step != stepInputName {
		t.Errorf("ESC 应该返回名称步骤，期望 stepInputName，实际 %d", m.step)
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

// TestBackspaceNotGoBack 测试 Backspace 不应返回上一步
func TestBackspaceNotGoBack(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 确认名称步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 输入 token 并经过 haiku/sonnet 到达 opus 步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test-token")})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputOpusModel {
		t.Fatalf("期望步骤为 stepInputOpusModel，实际为 %d", m.step)
	}

	// 清空 opus 输入框（通过多次按 Backspace）
	initialValue := m.opusInput.Value()
	for i := 0; i < len(initialValue)+10; i++ {
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = toModel(result)
	}

	// 验证：即使输入框为空，按 Backspace 也不应返回上一步
	if m.step != stepInputOpusModel {
		t.Errorf("Backspace 不应导致步骤改变。期望 stepInputOpusModel，实际 %d", m.step)
	}

	// 验证输入框确实为空
	if m.opusInput.Value() != "" {
		t.Errorf("输入框应为空，实际为 '%s'", m.opusInput.Value())
	}
}

// TestBackspaceNotDeletePreviousStep 测试 Backspace 不应删除上一步的内容
func TestBackspaceNotDeletePreviousStep(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 确认名称步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 输入 token
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("my-secret-token")})
	m = toModel(result)
	tokenValue := m.tokenInput.Value()

	// 进入 haiku/sonnet 步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputSonnetModel {
		t.Fatalf("期望步骤为 stepInputSonnetModel，实际为 %d", m.step)
	}

	// 在 sonnet 步骤多次按 Backspace（超过输入内容长度）
	for i := 0; i < 50; i++ {
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = toModel(result)
	}

	// 验证：token 值不应被修改
	if m.tokenInput.Value() != tokenValue {
		t.Errorf("Backspace 不应删除上一步的 token。期望 '%s'，实际 '%s'",
			tokenValue, m.tokenInput.Value())
	}
}

// TestPresetGoesToNameStep 测试选择预设后进入名称步骤（可编辑预设名）
func TestPresetGoesToNameStep(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputName {
		t.Fatalf("期望步骤为 stepInputName，实际为 %d", m.step)
	}

	// 验证名称已预填为预设名
	if m.nameInput.Value() != "anthropic" {
		t.Fatalf("期望名称预填为 'anthropic'，实际为 '%s'", m.nameInput.Value())
	}

	// 验证 presetName 已设置
	if m.presetName != "anthropic" {
		t.Fatalf("期望 presetName 为 'anthropic'，实际为 '%s'", m.presetName)
	}
}

// TestPresetCustomName 测试预设模式下可以修改名称
func TestPresetCustomName(t *testing.T) {
	m := InitialModel()

	// 选择 anthropic 预设
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	// 清空默认名称并输入自定义名称
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU}) // 清空
	m = toModel(result)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("anthropic-2")})
	m = toModel(result)

	if m.nameInput.Value() != "anthropic-2" {
		t.Fatalf("期望名称为 'anthropic-2'，实际为 '%s'", m.nameInput.Value())
	}

	// 确认名称，进入 token 步骤
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = toModel(result)

	if m.step != stepInputToken {
		t.Fatalf("期望步骤为 stepInputToken，实际为 %d", m.step)
	}
}
