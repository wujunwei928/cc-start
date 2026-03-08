// internal/repl/update.go
package repl

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/i18n"
	"github.com/wujunwei/cc-start/internal/theme"
	"github.com/wujunwei/cc-start/internal/tui/setup"
)

// Update 处理消息更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 设置面板激活时的处理
		if m.settings != nil && m.settings.IsVisible() {
			return m.updateSettings(msg)
		}

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
			if m.settings == nil {
				m.settings = NewSettingsPanel(m.styles, m.i18n)
			}
			m.settings.Toggle()
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

		case keyMatches(msg, m.keys.Esc):
			// 关闭面板（如果有打开的话）
			return m, nil

		default:
			// 检测 "/" 字符输入 - 打开命令面板
			if msg.String() == "/" && m.input.Value() == "" {
				if m.palette == nil {
					m.palette = NewCommandPalette(m.styles, m.i18n)
				}
				m.palette.Toggle()
				return m, nil
			}

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
		if m.settings != nil {
			m.settings.SetWidth(msg.Width)
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

	// 外部 TUI 完成后的配置重载消息
	case ConfigReloadMsg:
		cfg, err := config.LoadConfig(m.configPath)
		if err != nil {
			m.output.WriteError("重新加载配置失败: " + err.Error())
			return m, nil
		}
		m.config = cfg
		if msg.ProfileSaved != "" {
			m.output.WriteSuccess(fmt.Sprintf("配置 '%s' 已保存", msg.ProfileSaved))
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// ConfigReloadMsg 配置重载消息
type ConfigReloadMsg struct {
	ProfileSaved string
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

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		action := m.settings.SelectedAction()
		if action != "" {
			if m.settings.GetMode() != SettingsModeMain {
				result, cmd := m.handleSettingAction(action)
				m.settings.BackToMain()
				m.settings.Toggle()
				return result, cmd
			}
			m.settings.Toggle()
			return m.handleSettingAction(action)
		}
		return m, nil
	case "esc":
		if m.settings.GetMode() != SettingsModeMain {
			m.settings.BackToMain()
			return m, nil
		}
		m.settings.Toggle()
		return m, nil
	case "up", "down", "backspace":
		m.settings.HandleKey(msg.String())
		return m, nil
	default:
		if len(msg.Runes) > 0 {
			m.settings.HandleKey(string(msg.Runes))
		}
		return m, nil
	}
}

// handleSettingAction 处理设置动作
func (m Model) handleSettingAction(action string) (tea.Model, tea.Cmd) {
	if strings.HasPrefix(action, "lang:") {
		lang := strings.TrimPrefix(action, "lang:")
		return m.applyLanguageChange(lang)
	}
	if strings.HasPrefix(action, "theme:") {
		themeName := strings.TrimPrefix(action, "theme:")
		return m.applyThemeChange(themeName)
	}

	switch action {
	case "setting:lang":
		if m.settings != nil {
			m.settings.EnterSubMenu(SettingsModeLanguage)
			m.settings.visible = true
		}
		return m, nil
	case "setting:theme":
		if m.settings != nil {
			m.settings.EnterSubMenu(SettingsModeTheme)
			m.settings.visible = true
		}
		return m, nil
	default:
		m.output.Write("● 未知设置项: " + action)
	}
	return m, nil
}

// applyLanguageChange 应用语言更改
func (m *Model) applyLanguageChange(lang string) (tea.Model, tea.Cmd) {
	if err := m.i18n.SetLanguage(lang); err != nil {
		m.output.WriteError("不支持的语言: " + lang)
		return m, nil
	}

	m.config.Settings.Language = lang
	if err := m.config.Save(m.configPath); err != nil {
		m.output.WriteError("保存配置失败: " + err.Error())
		return m, nil
	}

	m.input.Placeholder = m.i18n.T(i18n.MsgREPLInputPrompt)
	if m.palette != nil {
		m.palette.SetI18n(m.i18n)
	}
	if m.settings != nil {
		m.settings.SetI18n(m.i18n)
	}

	m.output.WriteSuccess("语言已切换: " + lang)
	return m, nil
}

// applyThemeChange 应用主题更改
func (m *Model) applyThemeChange(themeName string) (tea.Model, tea.Cmd) {
	newTheme, err := theme.GetTheme(themeName)
	if err != nil {
		m.output.WriteError("不支持的主题: " + themeName)
		return m, nil
	}

	m.theme = newTheme
	m.styles = NewStylesFromTheme(newTheme)

	m.config.Settings.Theme = themeName
	if err := m.config.Save(m.configPath); err != nil {
		m.output.WriteError("保存配置失败: " + err.Error())
		return m, nil
	}

	if m.settings != nil {
		m.settings.SetStyles(m.styles)
	}
	if m.palette != nil {
		m.palette.SetStyles(m.styles)
	}

	m.output.WriteSuccess("主题已切换: " + themeName)
	return m, nil
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

	// 处理需要外部 TUI 的命令
	switch cmd {
	case "/setup":
		return m.runSetup(args)
	case "/edit":
		return m.runEdit(args)
	case "/run":
		return m.runLaunch(args)
	}

	// 收集输出
	output := m.collectCommandOutput(cmd, args)
	m.output.Write(output)
	return m, nil
}

// collectCommandOutput 执行命令并收集输出
func (m *Model) collectCommandOutput(cmd string, args []string) string {
	switch cmd {
	case "/list", "/ls":
		return m.formatProfileList()
	case "/use", "/switch":
		return m.cmdUse(args)
	case "/current", "/status":
		return m.formatCurrentProfile()
	case "/default":
		return m.cmdDefault(args)
	case "/show":
		return m.cmdShow(args)
	case "/add", "/new":
		return "请使用 '/setup' 命令进行交互式配置\n或使用 '/import' 命令导入配置文件"
	case "/delete", "/rm":
		return m.cmdDelete(args)
	case "/copy", "/cp":
		return m.cmdCopy(args)
	case "/rename", "/mv":
		return m.cmdRename(args)
	case "/test":
		return m.cmdTest(args)
	case "/export":
		return m.cmdExport(args)
	case "/import":
		return m.cmdImport(args)
	case "/history":
		return m.formatHistory()
	case "/help", "/?", "/h":
		return m.formatHelp()
	case "/clear", "/cls":
		m.output.Clear()
		return ""
	default:
		return fmt.Sprintf("未知命令: %s\n输入 '/help' 查看可用命令", cmd)
	}
}

// ========== 命令实现 ==========

// cmdUse 切换当前会话配置
func (m *Model) cmdUse(args []string) string {
	if len(args) == 0 {
		return "✗ 请指定配置名称: /use <name>"
	}

	name := args[0]
	profile, err := m.config.GetProfile(name)
	if err != nil {
		return "✗ " + err.Error()
	}

	m.currentProfile = profile.Name
	result := fmt.Sprintf("✓ 已切换到配置 '%s'", profile.Name)
	if profile.Model != "" {
		result += fmt.Sprintf("\n● 模型: %s", profile.Model)
	}
	return result
}

// cmdDefault 设置或显示默认配置
func (m *Model) cmdDefault(args []string) string {
	if len(args) == 0 {
		if m.config.Default == "" {
			return "⚠ 尚未设置默认配置"
		}
		return fmt.Sprintf("● 默认配置: %s", m.config.Default)
	}

	name := args[0]
	if err := m.config.SetDefault(name); err != nil {
		return "✗ " + err.Error()
	}

	if err := m.config.Save(m.configPath); err != nil {
		return "✗ 保存配置失败: " + err.Error()
	}

	return fmt.Sprintf("✓ 已将 '%s' 设为默认配置", name)
}

// cmdShow 显示配置详情
func (m *Model) cmdShow(args []string) string {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if m.currentProfile != "" {
		name = m.currentProfile
	}

	if name == "" {
		return "✗ 请指定配置名称: /show <name>"
	}

	profile, err := m.config.GetProfile(name)
	if err != nil {
		return "✗ " + err.Error()
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("\n配置名称: %s\n", profile.Name))
	buf.WriteString(fmt.Sprintf("Base URL: %s\n", profile.BaseURL))
	if profile.Model != "" {
		buf.WriteString(fmt.Sprintf("模型: %s\n", profile.Model))
	}
	buf.WriteString(fmt.Sprintf("Token: %s\n", maskAPIKey(profile.Token)))
	return buf.String()
}

// cmdDelete 删除配置
func (m *Model) cmdDelete(args []string) string {
	if len(args) == 0 {
		return "✗ 请指定配置名称: /delete <name> [-f]"
	}

	name := args[0]
	force := len(args) > 1 && args[1] == "-f"

	// 检查配置是否存在
	_, err := m.config.GetProfile(name)
	if err != nil {
		return "✗ " + err.Error()
	}

	// 在 TUI 环境中，删除需要 -f 参数确认
	if !force {
		return fmt.Sprintf("⚠ 确定要删除配置 '%s' 吗？\n使用 '/delete %s -f' 确认删除", name, name)
	}

	// 执行删除
	if err := m.config.DeleteProfile(name); err != nil {
		return "✗ " + err.Error()
	}

	// 如果删除的是当前配置，切换到默认配置
	if m.currentProfile == name {
		m.currentProfile = m.config.Default
	}

	if err := m.config.Save(m.configPath); err != nil {
		return "✗ 保存配置失败: " + err.Error()
	}

	return fmt.Sprintf("✓ 已删除配置 '%s'", name)
}

// cmdCopy 复制配置
func (m *Model) cmdCopy(args []string) string {
	if len(args) < 2 {
		return "✗ 用法: /copy <源配置> <新配置>"
	}

	srcName := args[0]
	dstName := args[1]

	src, err := m.config.GetProfile(srcName)
	if err != nil {
		return "✗ " + err.Error()
	}

	// 检查目标是否已存在
	for _, p := range m.config.Profiles {
		if p.Name == dstName {
			return fmt.Sprintf("✗ 配置 '%s' 已存在", dstName)
		}
	}

	// 创建新配置
	newProfile := config.Profile{
		Name:    dstName,
		BaseURL: src.BaseURL,
		Model:   src.Model,
		Token:   src.Token,
	}

	if err := m.config.AddProfile(newProfile); err != nil {
		return "✗ 添加配置失败: " + err.Error()
	}

	if err := m.config.Save(m.configPath); err != nil {
		return "✗ 保存配置失败: " + err.Error()
	}

	return fmt.Sprintf("✓ 已复制配置 '%s' -> '%s'", srcName, dstName)
}

// cmdRename 重命名配置
func (m *Model) cmdRename(args []string) string {
	if len(args) < 2 {
		return "✗ 用法: /rename <旧名称> <新名称>"
	}

	oldName := args[0]
	newName := args[1]

	profile, err := m.config.GetProfile(oldName)
	if err != nil {
		return "✗ " + err.Error()
	}

	// 检查新名称是否已存在
	for _, p := range m.config.Profiles {
		if p.Name == newName {
			return fmt.Sprintf("✗ 配置 '%s' 已存在", newName)
		}
	}

	// 更新配置名称
	profile.Name = newName

	// 如果重命名的是默认配置，更新默认名称
	if m.config.Default == oldName {
		m.config.Default = newName
	}

	// 如果重命名的是当前配置，更新当前名称
	if m.currentProfile == oldName {
		m.currentProfile = newName
	}

	// 删除旧配置，添加新配置
	m.config.DeleteProfile(oldName)
	m.config.AddProfile(*profile)

	if err := m.config.Save(m.configPath); err != nil {
		return "✗ 保存配置失败: " + err.Error()
	}

	return fmt.Sprintf("✓ 已重命名配置 '%s' -> '%s'", oldName, newName)
}

// cmdTest 测试 API 连通性
func (m *Model) cmdTest(args []string) string {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if m.currentProfile != "" {
		name = m.currentProfile
	}

	if name == "" {
		return "✗ 请指定配置名称: /test <name>"
	}

	profile, err := m.config.GetProfile(name)
	if err != nil {
		return "✗ " + err.Error()
	}

	result := fmt.Sprintf("● 测试配置 '%s' 的 API 连通性...\n", name)

	baseURL := profile.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	// 使用 curl 测试连接
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
		"--connect-timeout", "10",
		"-X", "GET",
		baseURL)
	output, err := cmd.Output()
	if err != nil {
		return result + "✗ 连接失败: " + err.Error()
	}

	statusCode := strings.TrimSpace(string(output))
	switch statusCode {
	case "200", "301", "302", "403", "404":
		result += fmt.Sprintf("✓ API 端点可达 (HTTP %s)\n● Base URL: %s", statusCode, baseURL)
	default:
		result += fmt.Sprintf("⚠ API 返回状态码: %s\n● 这可能表示服务不可用或网络问题", statusCode)
	}
	return result
}

// cmdExport 导出配置
func (m *Model) cmdExport(args []string) string {
	if len(args) == 0 {
		// 导出到 stdout
		data, err := json.MarshalIndent(m.config, "", "  ")
		if err != nil {
			return "✗ 序列化配置失败: " + err.Error()
		}
		return string(data)
	}

	// 导出到文件
	filePath := args[0]
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return "✗ 序列化配置失败: " + err.Error()
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "✗ 创建目录失败: " + err.Error()
		}
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "✗ 写入文件失败: " + err.Error()
	}

	return fmt.Sprintf("✓ 已导出配置到 '%s'", filePath)
}

// cmdImport 从文件导入配置
func (m *Model) cmdImport(args []string) string {
	if len(args) == 0 {
		return "✗ 请指定文件路径: /import <file>"
	}

	filePath := args[0]
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "✗ 读取文件失败: " + err.Error()
	}

	var importCfg config.Config
	if err := json.Unmarshal(data, &importCfg); err != nil {
		return "✗ 解析配置失败: " + err.Error()
	}

	if len(importCfg.Profiles) == 0 {
		return "⚠ 文件中没有找到配置"
	}

	// 导入配置，处理冲突
	imported := 0
	skipped := 0
	var skippedNames []string
	for _, p := range importCfg.Profiles {
		exists := false
		for _, existing := range m.config.Profiles {
			if existing.Name == p.Name {
				exists = true
				break
			}
		}

		if exists {
			skippedNames = append(skippedNames, p.Name)
			skipped++
			continue
		}

		if err := m.config.AddProfile(p); err != nil {
			skipped++
			continue
		}
		imported++
	}

	if err := m.config.Save(m.configPath); err != nil {
		return "✗ 保存配置失败: " + err.Error()
	}

	result := fmt.Sprintf("✓ 导入完成: %d 个配置已添加, %d 个已跳过", imported, skipped)
	if len(skippedNames) > 0 {
		result += fmt.Sprintf("\n● 跳过的配置: %s", strings.Join(skippedNames, ", "))
	}
	return result
}

// ========== TUI 命令 ==========

// runSetup 运行配置向导
func (m Model) runSetup(args []string) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		// 退出当前 TUI，启动 setup
		tea.Quit()
		setupModel := setup.InitialModel()
		p := tea.NewProgram(setupModel, tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			return ConfigReloadMsg{}
		}

		profileSaved := ""
		if final, ok := result.(setup.Model); ok && final.Done() {
			profileSaved = final.GetName()
		}
		return ConfigReloadMsg{ProfileSaved: profileSaved}
	}
}

// runEdit 编辑配置
func (m Model) runEdit(args []string) (tea.Model, tea.Cmd) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if m.currentProfile != "" {
		name = m.currentProfile
	}

	if name == "" {
		m.output.WriteError("请指定配置名称: /edit <name>")
		return m, nil
	}

	profile, err := m.config.GetProfile(name)
	if err != nil {
		m.output.WriteError(err.Error())
		return m, nil
	}

	return m, func() tea.Msg {
		tea.Quit()
		setupModel := setup.InitialModelWithProfile(*profile)
		p := tea.NewProgram(setupModel, tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			return ConfigReloadMsg{}
		}

		profileSaved := ""
		if final, ok := result.(setup.Model); ok && final.Done() {
			profileSaved = final.GetName()
		}
		return ConfigReloadMsg{ProfileSaved: profileSaved}
	}
}

// runLaunch 启动 Claude Code
func (m Model) runLaunch(args []string) (tea.Model, tea.Cmd) {
	profileName := m.currentProfile
	launchArgs := args

	// 解析参数
	for i, arg := range args {
		if arg == "--" {
			launchArgs = args[i+1:]
			if i > 0 {
				profileName = args[0]
			}
			break
		}
	}

	// 如果没有 -- 分隔符，检查第一个参数是否为 profile 名称
	if len(args) > 0 && len(launchArgs) == len(args) {
		if _, err := m.config.GetProfile(args[0]); err == nil {
			profileName = args[0]
			launchArgs = args[1:]
		}
	}

	if profileName == "" {
		m.output.WriteError("请先选择配置: /use <name>")
		m.output.WriteInfo("或指定配置名: /run <profile>")
		return m, nil
	}

	profile, err := m.config.GetProfile(profileName)
	if err != nil {
		m.output.WriteError("配置无效: " + err.Error())
		return m, nil
	}

	// 设置待执行的启动命令并退出
	m.PendingLaunch = &PendingLaunch{
		Profile: *profile,
		Args:    launchArgs,
	}
	m.quitting = true
	return m, tea.Quit
}

// ========== 格式化方法 ==========

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

// formatProfileList 格式化配置列表
func (m Model) formatProfileList() string {
	if len(m.config.Profiles) == 0 {
		return "⚠ 尚未配置任何供应商\n● 运行 '/setup' 创建配置"
	}

	var buf strings.Builder
	buf.WriteString("\n配置列表:\n")

	// 按名称排序输出
	names := make([]string, 0, len(m.config.Profiles))
	profileMap := make(map[string]config.Profile)
	for _, p := range m.config.Profiles {
		names = append(names, p.Name)
		profileMap[p.Name] = p
	}
	sort.Strings(names)

	for _, name := range names {
		p := profileMap[name]
		status := ""
		if name == m.config.Default {
			status = " [默认]"
		}
		if name == m.currentProfile {
			status += " [当前]"
		}
		buf.WriteString(fmt.Sprintf("  %s%s\n", p.Name, status))
		buf.WriteString(fmt.Sprintf("    Base URL: %s\n", p.BaseURL))
		if p.Model != "" {
			buf.WriteString(fmt.Sprintf("    模型: %s\n", p.Model))
		}
	}
	return buf.String()
}

// formatCurrentProfile 格式化当前配置
func (m Model) formatCurrentProfile() string {
	if m.currentProfile == "" {
		return "⚠ 当前未选择任何配置\n● 使用 '/use <name>' 选择配置"
	}

	profile, err := m.config.GetProfile(m.currentProfile)
	if err != nil {
		return "✗ 当前配置无效: " + err.Error()
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("\n当前配置: %s\n", profile.Name))
	buf.WriteString(fmt.Sprintf("  Base URL: %s\n", profile.BaseURL))
	if profile.Model != "" {
		buf.WriteString(fmt.Sprintf("  模型: %s\n", profile.Model))
	}
	buf.WriteString(fmt.Sprintf("  Token: %s\n", maskAPIKey(profile.Token)))
	if profile.Name == m.config.Default {
		buf.WriteString("● 这是默认配置\n")
	}
	return buf.String()
}

// formatHistory 格式化命令历史
func (m Model) formatHistory() string {
	commands := m.history.GetCommands()
	if len(commands) == 0 {
		return "● 暂无命令历史"
	}

	// 显示最近的 20 条命令
	start := 0
	if len(commands) > 20 {
		start = len(commands) - 20
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("\n● 最近 %d 条命令:\n", len(commands)-start))
	for i := start; i < len(commands); i++ {
		buf.WriteString(fmt.Sprintf("  %3d  %s\n", i+1, commands[i]))
	}
	return buf.String()
}

// formatHelp 格式化帮助
func (m Model) formatHelp() string {
	return `
可用命令:

配置管理:
  /list, /ls          列出所有配置
  /use, /switch       切换当前会话配置
  /current, /status   显示当前配置
  /default            设置默认配置
  /show               显示配置详情
  /add, /new          添加配置（提示使用 setup）
  /edit               编辑配置
  /delete, /rm        删除配置
  /copy, /cp          复制配置
  /rename, /mv        重命名配置

测试与导入导出:
  /test               测试 API 连通性
  /export             导出配置到 stdout 或文件
  /import             从文件导入配置

辅助命令:
  /history            显示命令历史
  /help, /?, /h       显示帮助
  /clear, /cls        清屏
  /exit, /quit, /q    退出

启动 Claude Code:
  /run [profile] [-- args...]  使用当前或指定配置启动
  /setup              运行配置向导
`
}
