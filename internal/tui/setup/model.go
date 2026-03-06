// internal/tui/setup/model.go
package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wujunwei/cc-start/internal/config"
)

// 步骤状态
type step int

const (
	stepSelectPreset step = iota
	stepInputName
	stepInputToken
	stepInputModel
	stepConfirm
	stepDone
)

// 样式定义
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Padding(1, 0)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
)

// Model setup TUI 模型
type Model struct {
	step       step
	presets    []string
	selected   int
	nameInput  textinput.Model
	tokenInput textinput.Model
	modelInput textinput.Model
	isCustom   bool
	presetName string
	baseURL    string
	err        error
	profile    *config.Profile
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

// Update 更新状态
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.step == stepSelectPreset && m.selected > 0 {
				m.selected--
			}
			return m, nil

		case tea.KeyDown:
			if m.step == stepSelectPreset && m.selected < len(m.presets)-1 {
				m.selected++
			}
			return m, nil

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyBackspace:
			return m.handleBackspace(msg)
		}
	}

	// 处理输入
	switch m.step {
	case stepInputName:
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	case stepInputToken:
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	case stepInputModel:
		var cmd tea.Cmd
		m.modelInput, cmd = m.modelInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepSelectPreset:
		m.presetName = m.presets[m.selected]
		if m.presetName == "自定义" {
			m.isCustom = true
			m.step = stepInputName
			m.nameInput.Focus()
		} else {
			// 使用预设
			preset, err := config.GetPresetByName(m.presetName)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.baseURL = preset.BaseURL
			m.modelInput.SetValue(preset.Model)
			m.nameInput.SetValue(preset.Name)
			m.step = stepInputToken
			m.tokenInput.Focus()
		}

	case stepInputName:
		if m.nameInput.Value() == "" {
			m.err = fmt.Errorf("配置名称不能为空")
			return m, nil
		}
		m.step = stepInputToken
		m.nameInput.Blur()
		m.tokenInput.Focus()

	case stepInputToken:
		if m.tokenInput.Value() == "" {
			m.err = fmt.Errorf("Token 不能为空")
			return m, nil
		}
		m.step = stepInputModel
		m.tokenInput.Blur()
		m.modelInput.Focus()

	case stepInputModel:
		m.saveProfile()
		return m, tea.Quit

	case stepConfirm:
		return m, tea.Quit
	}

	m.err = nil
	return m, nil
}

func (m *Model) handleBackspace(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 在输入步骤按 Backspace 可以返回上一步
	if m.step >= stepInputName && m.step < stepConfirm {
		// 检查输入框是否为空
		switch m.step {
		case stepInputName:
			if m.nameInput.Value() == "" {
				m.step = stepSelectPreset
				m.nameInput.Blur()
				return m, nil
			}
		case stepInputToken:
			if m.tokenInput.Value() == "" {
				if m.isCustom {
					m.step = stepInputName
					m.tokenInput.Blur()
					m.nameInput.Focus()
				} else {
					m.step = stepSelectPreset
					m.tokenInput.Blur()
				}
				return m, nil
			}
		case stepInputModel:
			if m.modelInput.Value() == "" {
				m.step = stepInputToken
				m.modelInput.Blur()
				m.tokenInput.Focus()
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *Model) saveProfile() {
	m.profile = &config.Profile{
		Name:    m.nameInput.Value(),
		BaseURL: m.baseURL,
		Token:   m.tokenInput.Value(),
		Model:   m.modelInput.Value(),
	}

	// 保存到文件
	cfgPath := config.GetConfigPath()
	cfg, _ := config.LoadConfig(cfgPath)
	cfg.AddProfile(*m.profile)

	// 如果是第一个配置，设为默认
	if len(cfg.Profiles) == 1 {
		cfg.Default = m.profile.Name
	}

	cfg.Save(cfgPath)
	m.step = stepDone
}

// View 渲染视图
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🚀 CC-Start 配置向导"))
	b.WriteString("\n\n")

	switch m.step {
	case stepSelectPreset:
		b.WriteString("选择预设:\n\n")
		for i, preset := range m.presets {
			if i == m.selected {
				b.WriteString(selectedStyle.Render("  → " + preset))
			} else {
				b.WriteString(normalStyle.Render("    " + preset))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(normalStyle.Render("↑/↓ 选择，Enter 确认"))

	case stepInputName:
		b.WriteString("输入配置名称:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.nameInput.View()))
		b.WriteString(normalStyle.Render("Enter 确认"))

	case stepInputToken:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		b.WriteString(fmt.Sprintf("URL: %s\n\n", m.baseURL))
		b.WriteString("输入 API Token:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.tokenInput.View()))
		b.WriteString(normalStyle.Render("Enter 确认"))

	case stepInputModel:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		b.WriteString(fmt.Sprintf("URL: %s\n\n", m.baseURL))
		b.WriteString("输入模型名称（可选）:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.modelInput.View()))
		b.WriteString(normalStyle.Render("Enter 保存，留空使用默认"))
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("❌ %v", m.err)))
	}

	return b.String()
}

// Done 返回是否完成
func (m Model) Done() bool {
	return m.step == stepDone
}

// GetName 返回配置名
func (m Model) GetName() string {
	return m.nameInput.Value()
}
