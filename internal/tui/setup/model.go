// internal/tui/setup/model.go
package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wujunwei928/cc-start/internal/config"
)

// 步骤状态
type step int

const (
	stepSelectPreset step = iota
	stepInputName
	stepInputAnthropicURL
	stepInputOpenAIURL
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
	// URL 输入
	anthropicURLInput textinput.Model
	openaiURLInput    textinput.Model
	isCustom          bool
	presetName        string
	err               error
	profile           *config.Profile
	// 编辑模式
	isEdit       bool
	originalName string // 原始配置名，用于重命名时更新引用
}

// InitialModel 创建初始模型
func InitialModel() Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "Profile name (e.g. my-api)"
	nameInput.Focus()

	tokenInput := textinput.New()
	tokenInput.Placeholder = "API Token"
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'

	modelInput := textinput.New()
	modelInput.Placeholder = "Model name (optional, press Enter to skip)"

	anthropicURLInput := textinput.New()
	anthropicURLInput.Placeholder = "Anthropic format URL (e.g. https://api.anthropic.com)"

	openaiURLInput := textinput.New()
	openaiURLInput.Placeholder = "OpenAI format URL (e.g. https://api.openai.com/v1)"

	return Model{
		step:              stepSelectPreset,
		presets:           []string{"anthropic", "moonshot", "bigmodel", "deepseek", "minimax", "自定义"},
		selected:          0,
		nameInput:         nameInput,
		tokenInput:        tokenInput,
		modelInput:        modelInput,
		anthropicURLInput: anthropicURLInput,
		openaiURLInput:    openaiURLInput,
	}
}

// InitialModelWithProfile 创建编辑模式的模型
func InitialModelWithProfile(p config.Profile) Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "Profile name"
	nameInput.SetValue(p.Name)
	nameInput.Focus()

	tokenInput := textinput.New()
	tokenInput.Placeholder = "API Token"
	tokenInput.SetValue(p.Token)
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'

	modelInput := textinput.New()
	modelInput.Placeholder = "Model name (optional)"
	modelInput.SetValue(p.Model)

	anthropicURLInput := textinput.New()
	anthropicURLInput.Placeholder = "Anthropic format URL"
	anthropicURLInput.SetValue(p.AnthropicBaseURL)

	openaiURLInput := textinput.New()
	openaiURLInput.Placeholder = "OpenAI format URL"
	openaiURLInput.SetValue(p.OpenAIBaseURL)

	// 查找匹配的预设
	presets := []string{"anthropic", "moonshot", "bigmodel", "deepseek", "minimax", "自定义"}
	selected := len(presets) - 1 // 默认选择"自定义"
	for i, preset := range presets[:len(presets)-1] {
		if presetConf, err := config.GetPresetByName(preset); err == nil {
			if presetConf.AnthropicBaseURL == p.AnthropicBaseURL {
				selected = i
				break
			}
		}
	}

	return Model{
		step:              stepInputName, // 编辑模式直接从名称输入开始
		presets:           presets,
		selected:          selected,
		nameInput:         nameInput,
		tokenInput:        tokenInput,
		modelInput:        modelInput,
		anthropicURLInput: anthropicURLInput,
		openaiURLInput:    openaiURLInput,
		isEdit:            true,
		originalName:      p.Name,
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
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			return m.handleGoBack()

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
	case stepInputAnthropicURL:
		var cmd tea.Cmd
		m.anthropicURLInput, cmd = m.anthropicURLInput.Update(msg)
		return m, cmd
	case stepInputOpenAIURL:
		var cmd tea.Cmd
		m.openaiURLInput, cmd = m.openaiURLInput.Update(msg)
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
			m.anthropicURLInput.SetValue(preset.AnthropicBaseURL)
			m.openaiURLInput.SetValue(preset.OpenAIBaseURL)
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
		// 编辑模式或自定义模式需要输入 URL
		if m.isEdit || m.isCustom {
			m.step = stepInputAnthropicURL
			m.nameInput.Blur()
			m.anthropicURLInput.Focus()
		} else {
			m.step = stepInputToken
			m.nameInput.Blur()
			m.tokenInput.Focus()
		}

	case stepInputAnthropicURL:
		// Anthropic URL 可以为空，继续下一步
		m.step = stepInputOpenAIURL
		m.anthropicURLInput.Blur()
		m.openaiURLInput.Focus()

	case stepInputOpenAIURL:
		// OpenAI URL 可以为空，继续下一步
		m.step = stepInputToken
		m.openaiURLInput.Blur()
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

func (m *Model) handleGoBack() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepInputModel:
		m.step = stepInputToken
		m.modelInput.Blur()
		m.tokenInput.Focus()
	case stepInputToken:
		if m.isEdit || m.isCustom {
			m.step = stepInputOpenAIURL
			m.tokenInput.Blur()
			m.openaiURLInput.Focus()
		} else {
			m.step = stepSelectPreset
			m.tokenInput.Blur()
		}
	case stepInputOpenAIURL:
		m.step = stepInputAnthropicURL
		m.openaiURLInput.Blur()
		m.anthropicURLInput.Focus()
	case stepInputAnthropicURL:
		m.step = stepInputName
		m.anthropicURLInput.Blur()
		m.nameInput.Focus()
	case stepInputName:
		if m.isEdit {
			return m, tea.Quit
		}
		m.step = stepSelectPreset
		m.nameInput.Blur()
	}
	m.err = nil
	return m, nil
}

func (m *Model) handleBackspace(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Backspace 只删除当前步骤的字符
	switch m.step {
	case stepInputName:
		if m.nameInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	case stepInputAnthropicURL:
		if m.anthropicURLInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.anthropicURLInput, cmd = m.anthropicURLInput.Update(msg)
		return m, cmd
	case stepInputOpenAIURL:
		if m.openaiURLInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.openaiURLInput, cmd = m.openaiURLInput.Update(msg)
		return m, cmd
	case stepInputToken:
		if m.tokenInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	case stepInputModel:
		if m.modelInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.modelInput, cmd = m.modelInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) saveProfile() {
	m.profile = &config.Profile{
		Name:             m.nameInput.Value(),
		AnthropicBaseURL: m.anthropicURLInput.Value(),
		OpenAIBaseURL:    m.openaiURLInput.Value(),
		Token:            m.tokenInput.Value(),
		Model:            m.modelInput.Value(),
	}

	// 保存到文件
	cfgPath := config.GetConfigPath()
	cfg, _ := config.LoadConfig(cfgPath)

	if m.isEdit && m.originalName != "" && m.originalName != m.profile.Name {
		// 编辑模式且名称改变：需要删除旧配置
		cfg.DeleteProfile(m.originalName)
		// 如果旧名称是默认配置，更新默认名称
		if cfg.Default == m.originalName {
			cfg.Default = m.profile.Name
		}
	}

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

	if m.isEdit {
		b.WriteString(titleStyle.Render("✏️ 编辑配置"))
	} else {
		b.WriteString(titleStyle.Render("🚀 CC-Start 配置向导"))
	}
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
		b.WriteString(normalStyle.Render("Enter 确认，ESC 返回"))

	case stepInputAnthropicURL:
		b.WriteString(fmt.Sprintf("配置: %s\n\n", m.nameInput.Value()))
		b.WriteString("输入 Anthropic 格式 URL（可留空）:\n")
		b.WriteString(normalStyle.Render("  用于 Claude CLI"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.anthropicURLInput.View()))
		b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

	case stepInputOpenAIURL:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		b.WriteString(fmt.Sprintf("Anthropic URL: %s\n\n", m.anthropicURLInput.Value()))
		b.WriteString("输入 OpenAI 格式 URL（可留空）:\n")
		b.WriteString(normalStyle.Render("  用于 Codex/OpenCode CLI"))
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.openaiURLInput.View()))
		b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

	case stepInputToken:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		if m.anthropicURLInput.Value() != "" {
			b.WriteString(fmt.Sprintf("Anthropic URL: %s\n", m.anthropicURLInput.Value()))
		}
		if m.openaiURLInput.Value() != "" {
			b.WriteString(fmt.Sprintf("OpenAI URL: %s\n", m.openaiURLInput.Value()))
		}
		b.WriteString("\n输入 API Token:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.tokenInput.View()))
		b.WriteString(normalStyle.Render("Enter 确认，ESC 返回"))

	case stepInputModel:
		b.WriteString(fmt.Sprintf("配置: %s\n", m.nameInput.Value()))
		if m.anthropicURLInput.Value() != "" {
			b.WriteString(fmt.Sprintf("Anthropic URL: %s\n", m.anthropicURLInput.Value()))
		}
		if m.openaiURLInput.Value() != "" {
			b.WriteString(fmt.Sprintf("OpenAI URL: %s\n", m.openaiURLInput.Value()))
		}
		b.WriteString("\n输入模型名称（可选）:\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.modelInput.View()))
		b.WriteString(normalStyle.Render("Enter 保存，ESC 返回"))
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
