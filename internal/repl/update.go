// internal/repl/update.go
package repl

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Update 处理消息更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 命令面板激活时的处理
		if m.palette != nil && m.palette.IsVisible() {
			return m.updatePalette(msg)
		}

		// 主界面按键处理
		switch {
		case keyMatches(msg, m.keys.CtrlC):
			m.quitting = true
			return m, tea.Quit

		case keyMatches(msg, m.keys.CtrlP):
			if m.palette == nil {
				m.palette = NewCommandPalette(m.styles)
			}
			m.palette.Toggle()
			return m, nil

		case keyMatches(msg, m.keys.CtrlL):
			m.output.Clear()
			return m, nil

		case keyMatches(msg, m.keys.Enter):
			return m.executeInput()

		case keyMatches(msg, m.keys.Up):
			return m.navigateHistory(-1)

		case keyMatches(msg, m.keys.Down):
			return m.navigateHistory(1)

		default:
			// 更新输入框
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		if m.palette != nil {
			m.palette.SetWidth(msg.Width)
		}
		return m, nil

	case CommandSelectedMsg:
		return m.executeCommand(msg.Cmd, msg.Args)

	case CommandExecutedMsg:
		if msg.Err != nil {
			m.output.WriteError(msg.Err.Error())
		} else {
			m.output.Write(msg.Output)
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updatePalette(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := m.palette.SelectedCommand()
		m.palette.Toggle()
		if cmd != "" {
			return m.executeCommand(cmd, nil)
		}
		return m, nil
	case "esc":
		m.palette.Toggle()
		return m, nil
	case "up", "down", "backspace":
		m.palette.HandleKey(msg.String())
		return m, nil
	default:
		// 字符输入
		if len(msg.Runes) > 0 {
			m.palette.HandleKey(string(msg.Runes))
		}
		return m, nil
	}
}

func (m Model) executeInput() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	if input == "" {
		return m, nil
	}

	m.history.Add(input)
	m.input.SetValue("")
	m.histIdx = 0

	// 解析命令
	parts := strings.Fields(input)
	cmd := parts[0]
	args := parts[1:]

	// 自动添加 / 前缀
	if !strings.HasPrefix(cmd, "/") {
		cmd = "/" + cmd
	}

	return m.executeCommand(cmd, args)
}

func (m Model) executeCommand(cmd string, args []string) (tea.Model, tea.Cmd) {
	// 检查退出命令
	switch cmd {
	case "/exit", "/quit", "/q":
		m.quitting = true
		return m, tea.Quit
	}

	// 收集输出
	output := m.collectCommandOutput(cmd, args)
	m.output.Write(output)
	return m, nil
}

// collectCommandOutput 执行命令并收集输出
func (m *Model) collectCommandOutput(cmd string, args []string) string {
	// 使用 strings.Builder 收集输出
	var buf strings.Builder

	// 执行命令（输出会被打印到 stdout）
	// 这里我们需要捕获输出
	// 暂时使用简化的方式
	switch cmd {
	case "/list", "/ls":
		buf.WriteString(m.formatProfileList())
	case "/current", "/status":
		buf.WriteString(m.formatCurrentProfile())
	case "/help", "/?", "/h":
		buf.WriteString(m.formatHelp())
	case "/clear", "/cls":
		m.output.Clear()
	default:
		// 其他命令暂时返回提示
		buf.WriteString(fmt.Sprintf("命令 %s 暂未实现\n", cmd))
	}

	return buf.String()
}

func (m Model) navigateHistory(dir int) (tea.Model, tea.Cmd) {
	cmds := m.history.GetCommands()
	if len(cmds) == 0 {
		return m, nil
	}

	newIdx := m.histIdx + dir
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx > len(cmds) {
		newIdx = len(cmds)
	}
	m.histIdx = newIdx

	if newIdx == 0 {
		m.input.SetValue("")
	} else {
		m.input.SetValue(cmds[newIdx-1])
	}
	m.input.CursorEnd()

	return m, nil
}

func keyMatches(msg tea.KeyMsg, binding interface{}) bool {
	b, ok := binding.(interface {
		Keys() []string
		Enabled() bool
	})
	if !ok {
		return false
	}
	if !b.Enabled() {
		return false
	}
	for _, k := range b.Keys() {
		if msg.String() == k {
			return true
		}
	}
	return false
}

// 格式化辅助方法
func (m Model) formatProfileList() string {
	if len(m.config.Profiles) == 0 {
		return "尚未配置任何供应商\n运行 '/setup' 创建配置"
	}

	var buf strings.Builder
	buf.WriteString("\n配置列表:\n")
	for _, p := range m.config.Profiles {
		status := ""
		if p.Name == m.config.Default {
			status = " [默认]"
		}
		if p.Name == m.currentProfile {
			status += " [当前]"
		}
		buf.WriteString(fmt.Sprintf("  %s%s\n", p.Name, status))
	}
	return buf.String()
}

func (m Model) formatCurrentProfile() string {
	if m.currentProfile == "" {
		return "当前未选择任何配置\n使用 '/use <name>' 选择配置"
	}

	profile, err := m.config.GetProfile(m.currentProfile)
	if err != nil {
		return "当前配置无效: " + err.Error()
	}

	var buf strings.Builder
	buf.WriteString("\n当前配置: " + profile.Name + "\n")
	buf.WriteString("  Base URL: " + profile.BaseURL + "\n")
	if profile.Model != "" {
		buf.WriteString("  模型: " + profile.Model + "\n")
	}
	return buf.String()
}

func (m Model) formatHelp() string {
	return `
可用命令:

配置管理:
  /list, /ls          列出所有配置
  /use, /switch       切换当前会话配置
  /current, /status   显示当前配置
  /default            设置默认配置
  /show               显示配置详情
  /edit               编辑配置
  /delete, /rm        删除配置

辅助命令:
  /history            显示命令历史
  /help, /?, /h       显示帮助
  /clear, /cls        清屏
  /exit, /quit, /q    退出

启动:
  /run [profile]      启动 Claude Code
  /setup              运行配置向导
`
}
