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
	stepInputToken
	stepInputHaikuModel
	stepInputSonnetModel
	stepInputOpusModel
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
	step        step
	presets     []string
	selected    int
	nameInput   textinput.Model
	tokenInput  textinput.Model
	haikuInput  textinput.Model
	sonnetInput textinput.Model
	opusInput   textinput.Model
	// URL 输入
	anthropicURLInput textinput.Model
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

	haikuInput := textinput.New()
	haikuInput.Placeholder = "快速模型 (Haiku)"

	sonnetInput := textinput.New()
	sonnetInput.Placeholder = "主模型 (Sonnet)"

	opusInput := textinput.New()
	opusInput.Placeholder = "经济模型 (Opus)"

	anthropicURLInput := textinput.New()
	anthropicURLInput.Placeholder = "Base URL (e.g. https://api.anthropic.com)"

	return Model{
		step:              stepSelectPreset,
		presets:           []string{"anthropic", "moonshot", "bigmodel", "deepseek", "minimax", "自定义"},
		selected:          0,
		nameInput:         nameInput,
		tokenInput:        tokenInput,
		haikuInput:        haikuInput,
		sonnetInput:       sonnetInput,
		opusInput:         opusInput,
		anthropicURLInput: anthropicURLInput,
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

	haikuInput := textinput.New()
	haikuInput.Placeholder = "快速模型 (Haiku)"
	haikuInput.SetValue(p.HaikuModel)

	sonnetInput := textinput.New()
	sonnetInput.Placeholder = "主模型 (Sonnet)"
	sonnetInput.SetValue(p.SonnetModel)

	opusInput := textinput.New()
	opusInput.Placeholder = "经济模型 (Opus)"
	opusInput.SetValue(p.OpusModel)

	anthropicURLInput := textinput.New()
	anthropicURLInput.Placeholder = "Base URL"
	anthropicURLInput.SetValue(p.AnthropicBaseURL)

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
		haikuInput:        haikuInput,
		sonnetInput:       sonnetInput,
		opusInput:         opusInput,
		anthropicURLInput: anthropicURLInput,
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
	case stepInputToken:
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	case stepInputHaikuModel:
		var cmd tea.Cmd
		m.haikuInput, cmd = m.haikuInput.Update(msg)
		return m, cmd
	case stepInputSonnetModel:
		var cmd tea.Cmd
		m.sonnetInput, cmd = m.sonnetInput.Update(msg)
		return m, cmd
	case stepInputOpusModel:
		var cmd tea.Cmd
		m.opusInput, cmd = m.opusInput.Update(msg)
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
			preset, err := config.GetPresetByName(m.presetName)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.anthropicURLInput.SetValue(preset.AnthropicBaseURL)
			m.haikuInput.SetValue(preset.HaikuModel)
			m.sonnetInput.SetValue(preset.SonnetModel)
			m.opusInput.SetValue(preset.OpusModel)
			m.nameInput.SetValue(preset.Name)
			m.step = stepInputName
			m.nameInput.Focus()
		}

	case stepInputName:
		if m.nameInput.Value() == "" {
			m.err = fmt.Errorf("配置名称不能为空")
			return m, nil
		}
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
		m.step = stepInputToken
		m.anthropicURLInput.Blur()
		m.tokenInput.Focus()

	case stepInputToken:
		if m.tokenInput.Value() == "" {
			m.err = fmt.Errorf("Token 不能为空")
			return m, nil
		}
		m.step = stepInputHaikuModel
		m.tokenInput.Blur()
		m.haikuInput.Focus()

	case stepInputHaikuModel:
		m.step = stepInputSonnetModel
		m.haikuInput.Blur()
		m.sonnetInput.Focus()

	case stepInputSonnetModel:
		m.step = stepInputOpusModel
		m.sonnetInput.Blur()
		m.opusInput.Focus()

	case stepInputOpusModel:
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
	case stepInputOpusModel:
		m.step = stepInputSonnetModel
		m.opusInput.Blur()
		m.sonnetInput.Focus()
	case stepInputSonnetModel:
		m.step = stepInputHaikuModel
		m.sonnetInput.Blur()
		m.haikuInput.Focus()
	case stepInputHaikuModel:
		m.step = stepInputToken
		m.haikuInput.Blur()
		m.tokenInput.Focus()
	case stepInputToken:
		if m.isEdit || m.isCustom {
			m.step = stepInputAnthropicURL
			m.tokenInput.Blur()
			m.anthropicURLInput.Focus()
		} else {
			m.step = stepInputName
			m.tokenInput.Blur()
			m.nameInput.Focus()
		}
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
	case stepInputToken:
		if m.tokenInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	case stepInputHaikuModel:
		if m.haikuInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.haikuInput, cmd = m.haikuInput.Update(msg)
		return m, cmd
	case stepInputSonnetModel:
		if m.sonnetInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.sonnetInput, cmd = m.sonnetInput.Update(msg)
		return m, cmd
	case stepInputOpusModel:
		if m.opusInput.Value() == "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.opusInput, cmd = m.opusInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) saveProfile() {
	haiku := m.haikuInput.Value()
	sonnet := m.sonnetInput.Value()
	opus := m.opusInput.Value()

	// 创建模式：留空使用预设默认值
	if !m.isEdit {
		if haiku == "" && m.presetName != "" {
			if preset, err := config.GetPresetByName(m.presetName); err == nil {
				haiku = preset.HaikuModel
			}
		}
		if sonnet == "" && m.presetName != "" {
			if preset, err := config.GetPresetByName(m.presetName); err == nil {
				sonnet = preset.SonnetModel
			}
		}
		if opus == "" && m.presetName != "" {
			if preset, err := config.GetPresetByName(m.presetName); err == nil {
				opus = preset.OpusModel
			}
		}
	}
	// 编辑模式：直接使用输入值（包括空字符串，表示清空）

	m.profile = &config.Profile{
		Name:             m.nameInput.Value(),
		AnthropicBaseURL: m.anthropicURLInput.Value(),
		HaikuModel:       haiku,
		SonnetModel:      sonnet,
		OpusModel:        opus,
		Token:            m.tokenInput.Value(),
	}

	cfgPath := config.GetConfigPath()
	cfg, _ := config.LoadConfig(cfgPath)

	if m.isEdit && m.originalName != "" && m.originalName != m.profile.Name {
		cfg.DeleteProfile(m.originalName)
		if cfg.Default == m.originalName {
			cfg.Default = m.profile.Name
		}
	}

	cfg.AddProfile(*m.profile)

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
		if !m.isCustom && m.presetName != "" {
			fmt.Fprintf(&b, "输入配置名称（预设: %s）:\n\n", m.presetName)
		} else {
			b.WriteString("输入配置名称:\n\n")
		}
		fmt.Fprintf(&b, "  %s\n\n", m.nameInput.View())
		b.WriteString(normalStyle.Render("Enter 确认，ESC 返回"))

	case stepInputAnthropicURL:
		fmt.Fprintf(&b, "配置: %s\n\n", m.nameInput.Value())
		b.WriteString("输入 Base URL（可留空）:\n")
		b.WriteString(normalStyle.Render("  用于 Claude CLI"))
		b.WriteString("\n\n")
		fmt.Fprintf(&b, "  %s\n\n", m.anthropicURLInput.View())
		b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

	case stepInputToken:
		fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
		if m.anthropicURLInput.Value() != "" {
			fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
		}
		b.WriteString("\n输入 API Token:\n\n")
		fmt.Fprintf(&b, "  %s\n\n", m.tokenInput.View())
		b.WriteString(normalStyle.Render("Enter 确认，ESC 返回"))

	case stepInputHaikuModel:
		fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
		if m.anthropicURLInput.Value() != "" {
			fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
		}
		b.WriteString("\n输入快速模型 (Haiku):\n\n")
		fmt.Fprintf(&b, "  %s\n\n", m.haikuInput.View())
		b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

	case stepInputSonnetModel:
		fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
		if m.anthropicURLInput.Value() != "" {
			fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
		}
		b.WriteString("\n输入主模型 (Sonnet):\n\n")
		fmt.Fprintf(&b, "  %s\n\n", m.sonnetInput.View())
		b.WriteString(normalStyle.Render("Enter 继续，ESC 返回"))

	case stepInputOpusModel:
		fmt.Fprintf(&b, "配置: %s\n", m.nameInput.Value())
		if m.anthropicURLInput.Value() != "" {
			fmt.Fprintf(&b, "Base URL: %s\n", m.anthropicURLInput.Value())
		}
		b.WriteString("\n输入经济模型 (Opus):\n\n")
		fmt.Fprintf(&b, "  %s\n\n", m.opusInput.View())
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
