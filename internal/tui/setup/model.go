// internal/tui/setup/model.go
package setup

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// 步骤状态
type step int

const (
	stepSelectPreset step = iota
	stepInputName
	stepInputToken
	stepInputModel
	stepConfirm
)

// Model setup TUI 模型
type Model struct {
	step        step
	presets     []string
	selected    int
	nameInput   textinput.Model
	tokenInput  textinput.Model
	modelInput  textinput.Model
	isCustom    bool
	err         error
}

// InitialModel 创建初始模型
func InitialModel() Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "配置名称（如 my-api）"
	nameInput.Focus()

	tokenInput := textinput.New()
	tokenInput.Placeholder = "API Token"
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'

	modelInput := textinput.New()
	modelInput.Placeholder = "模型名称（可选，按回车跳过）"

	return Model{
		step:       stepSelectPreset,
		presets:    []string{"anthropic", "moonshot", "bigmodel", "deepseek", "自定义"},
		selected:   0,
		nameInput:  nameInput,
		tokenInput: tokenInput,
		modelInput: modelInput,
	}
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 处理消息（后续实现完整逻辑）
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View 渲染视图（后续实现完整逻辑）
func (m Model) View() string {
	return "Setup TUI (待实现)\n\n按 Ctrl+C 退出"
}

// Done 返回是否完成
func (m Model) Done() bool {
	return false // 后续实现
}

// GetName 返回配置名
func (m Model) GetName() string {
	return m.nameInput.Value()
}
