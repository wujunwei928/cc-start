// internal/repl/commands.go
package repl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/launcher"
	"github.com/wujunwei/cc-start/internal/tui/setup"
)

// maskAPIKey 遮蔽 API Key 显示
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// cmdList 列出所有配置
func (r *REPL) cmdList(args []string) {
	if len(r.cfg.Profiles) == 0 {
		PrintWarning("尚未配置任何供应商")
		PrintInfo("运行 'setup' 创建配置")
		return
	}

	table := NewTable()
	table.Header([]string{"名称", "Base URL", "模型", "Token", "状态"})

	// 按名称排序输出
	names := make([]string, 0, len(r.cfg.Profiles))
	profileMap := make(map[string]config.Profile)
	for _, p := range r.cfg.Profiles {
		names = append(names, p.Name)
		profileMap[p.Name] = p
	}
	sort.Strings(names)

	for _, name := range names {
		p := profileMap[name]
		status := ""
		if name == r.cfg.Default {
			status = "默认"
		}
		if name == r.currentName {
			status += " 当前"
		}
		status = strings.TrimSpace(status)

		table.Append([]string{
			name,
			p.BaseURL,
			p.Model,
			maskAPIKey(p.Token),
			status,
		})
	}

	fmt.Println()
	table.Render()
	fmt.Println()
}

// cmdUse 切换当前会话配置
func (r *REPL) cmdUse(args []string) {
	if len(args) == 0 {
		PrintError("请指定配置名称: use <name>")
		return
	}

	name := args[0]
	profile, err := r.cfg.GetProfile(name)
	if err != nil {
		PrintError("%v", err)
		return
	}

	r.currentName = profile.Name
	PrintSuccess("已切换到配置 '%s'", profile.Name)
	if profile.Model != "" {
		PrintInfo("模型: %s", profile.Model)
	}
}

// cmdCurrent 显示当前配置
func (r *REPL) cmdCurrent(args []string) {
	if r.currentName == "" {
		PrintWarning("当前未选择任何配置")
		PrintInfo("使用 'use <name>' 选择配置")
		return
	}

	profile, err := r.cfg.GetProfile(r.currentName)
	if err != nil {
		PrintError("当前配置无效: %v", err)
		return
	}

	fmt.Println()
	PrintCurrent("当前配置: %s", profile.Name)
	fmt.Printf("  Base URL: %s\n", profile.BaseURL)
	if profile.Model != "" {
		fmt.Printf("  模型: %s\n", profile.Model)
	}
	fmt.Printf("  Token: %s\n", maskAPIKey(profile.Token))
	if profile.Name == r.cfg.Default {
		PrintInfo("这是默认配置")
	}
	fmt.Println()
}

// cmdDefault 设置默认配置
func (r *REPL) cmdDefault(args []string) {
	if len(args) == 0 {
		// 显示当前默认配置
		if r.cfg.Default == "" {
			PrintWarning("尚未设置默认配置")
		} else {
			PrintInfo("默认配置: %s", r.cfg.Default)
		}
		return
	}

	name := args[0]
	if err := r.cfg.SetDefault(name); err != nil {
		PrintError("%v", err)
		return
	}

	if err := r.cfg.Save(r.cfgPath); err != nil {
		PrintError("保存配置失败: %v", err)
		return
	}

	PrintSuccess("已将 '%s' 设为默认配置", name)
}

// cmdShow 显示配置详情
func (r *REPL) cmdShow(args []string) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if r.currentName != "" {
		name = r.currentName
	}

	if name == "" {
		PrintError("请指定配置名称: show <name>")
		return
	}

	profile, err := r.cfg.GetProfile(name)
	if err != nil {
		PrintError("%v", err)
		return
	}

	fmt.Println()
	fmt.Printf("配置名称: %s\n", profile.Name)
	fmt.Printf("Base URL: %s\n", profile.BaseURL)
	if profile.Model != "" {
		fmt.Printf("模型: %s\n", profile.Model)
	}
	fmt.Printf("Token: %s\n", maskAPIKey(profile.Token))
	fmt.Println()
}

// cmdAdd 添加配置（提示使用 setup）
func (r *REPL) cmdAdd(args []string) {
	PrintInfo("请使用 'setup' 命令进行交互式配置")
	PrintInfo("或使用 'import' 命令导入配置文件")
}

// cmdEdit 编辑配置
func (r *REPL) cmdEdit(args []string) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if r.currentName != "" {
		name = r.currentName
	}

	if name == "" {
		PrintError("请指定配置名称: edit <name>")
		return
	}

	profile, err := r.cfg.GetProfile(name)
	if err != nil {
		PrintError("%v", err)
		return
	}

	// 使用编辑模式启动 TUI
	m := setup.InitialModelWithProfile(*profile)
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		PrintError("启动 TUI 失败: %v", err)
		return
	}

	// 重新加载配置
	cfg, err := config.LoadConfig(r.cfgPath)
	if err != nil {
		PrintError("重新加载配置失败: %v", err)
		return
	}
	r.cfg = cfg

	if final, ok := result.(setup.Model); ok && final.Done() {
		PrintSuccess("配置 '%s' 已更新", final.GetName())
		// 如果编辑的是当前配置，更新 currentName
		if name != final.GetName() {
			r.currentName = final.GetName()
		}
	}
}

// cmdDelete 删除配置
func (r *REPL) cmdDelete(args []string) {
	if len(args) == 0 {
		PrintError("请指定配置名称: delete <name>")
		return
	}

	name := args[0]

	// 检查配置是否存在
	_, err := r.cfg.GetProfile(name)
	if err != nil {
		PrintError("%v", err)
		return
	}

	// 确认删除
	fmt.Printf("确定要删除配置 '%s' 吗? [y/N]: ", name)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		PrintInfo("已取消删除")
		return
	}

	if err := r.cfg.DeleteProfile(name); err != nil {
		PrintError("%v", err)
		return
	}

	// 如果删除的是当前配置，清除当前选择
	if r.currentName == name {
		r.currentName = r.cfg.Default
	}

	if err := r.cfg.Save(r.cfgPath); err != nil {
		PrintError("保存配置失败: %v", err)
		return
	}

	PrintSuccess("已删除配置 '%s'", name)
}

// cmdCopy 复制配置
func (r *REPL) cmdCopy(args []string) {
	if len(args) < 2 {
		PrintError("用法: copy <源配置> <新配置>")
		return
	}

	srcName := args[0]
	dstName := args[1]

	src, err := r.cfg.GetProfile(srcName)
	if err != nil {
		PrintError("%v", err)
		return
	}

	// 检查目标是否已存在
	for _, p := range r.cfg.Profiles {
		if p.Name == dstName {
			PrintError("配置 '%s' 已存在", dstName)
			return
		}
	}

	// 创建新配置
	newProfile := config.Profile{
		Name:    dstName,
		BaseURL: src.BaseURL,
		Model:   src.Model,
		Token:   src.Token,
	}

	if err := r.cfg.AddProfile(newProfile); err != nil {
		PrintError("添加配置失败: %v", err)
		return
	}

	if err := r.cfg.Save(r.cfgPath); err != nil {
		PrintError("保存配置失败: %v", err)
		return
	}

	PrintSuccess("已复制配置 '%s' -> '%s'", srcName, dstName)
}

// cmdRename 重命名配置
func (r *REPL) cmdRename(args []string) {
	if len(args) < 2 {
		PrintError("用法: rename <旧名称> <新名称>")
		return
	}

	oldName := args[0]
	newName := args[1]

	profile, err := r.cfg.GetProfile(oldName)
	if err != nil {
		PrintError("%v", err)
		return
	}

	// 检查新名称是否已存在
	for _, p := range r.cfg.Profiles {
		if p.Name == newName {
			PrintError("配置 '%s' 已存在", newName)
			return
		}
	}

	// 更新配置名称
	profile.Name = newName

	// 如果重命名的是默认配置，更新默认名称
	if r.cfg.Default == oldName {
		r.cfg.Default = newName
	}

	// 如果重命名的是当前配置，更新当前名称
	if r.currentName == oldName {
		r.currentName = newName
	}

	// 删除旧配置，添加新配置
	r.cfg.DeleteProfile(oldName)
	r.cfg.AddProfile(*profile)

	if err := r.cfg.Save(r.cfgPath); err != nil {
		PrintError("保存配置失败: %v", err)
		return
	}

	PrintSuccess("已重命名配置 '%s' -> '%s'", oldName, newName)
}

// cmdTest 测试 API 连通性
func (r *REPL) cmdTest(args []string) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if r.currentName != "" {
		name = r.currentName
	}

	if name == "" {
		PrintError("请指定配置名称: test <name>")
		return
	}

	profile, err := r.cfg.GetProfile(name)
	if err != nil {
		PrintError("%v", err)
		return
	}

	PrintInfo("测试配置 '%s' 的 API 连通性...", name)

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
		PrintError("连接失败: %v", err)
		return
	}

	statusCode := strings.TrimSpace(string(output))
	switch statusCode {
	case "200", "301", "302", "403", "404":
		// 这些状态码说明服务器可达（即使返回错误页面）
		PrintSuccess("API 端点可达 (HTTP %s)", statusCode)
		PrintInfo("Base URL: %s", baseURL)
	default:
		PrintWarning("API 返回状态码: %s", statusCode)
		PrintInfo("这可能表示服务不可用或网络问题")
	}
}

// cmdExport 导出配置
func (r *REPL) cmdExport(args []string) {
	if len(args) == 0 {
		// 导出到 stdout
		data, err := json.MarshalIndent(r.cfg, "", "  ")
		if err != nil {
			PrintError("序列化配置失败: %v", err)
			return
		}
		fmt.Println(string(data))
		return
	}

	// 导出到文件
	filePath := args[0]
	data, err := json.MarshalIndent(r.cfg, "", "  ")
	if err != nil {
		PrintError("序列化配置失败: %v", err)
		return
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			PrintError("创建目录失败: %v", err)
			return
		}
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		PrintError("写入文件失败: %v", err)
		return
	}

	PrintSuccess("已导出配置到 '%s'", filePath)
}

// cmdImport 从文件导入配置
func (r *REPL) cmdImport(args []string) {
	if len(args) == 0 {
		PrintError("请指定文件路径: import <file>")
		return
	}

	filePath := args[0]
	data, err := os.ReadFile(filePath)
	if err != nil {
		PrintError("读取文件失败: %v", err)
		return
	}

	var importCfg config.Config
	if err := json.Unmarshal(data, &importCfg); err != nil {
		PrintError("解析配置失败: %v", err)
		return
	}

	if len(importCfg.Profiles) == 0 {
		PrintWarning("文件中没有找到配置")
		return
	}

	// 导入配置，处理冲突
	imported := 0
	skipped := 0
	for _, p := range importCfg.Profiles {
		exists := false
		for _, existing := range r.cfg.Profiles {
			if existing.Name == p.Name {
				exists = true
				break
			}
		}

		if exists {
			fmt.Printf("配置 '%s' 已存在，跳过 (使用 'delete' 删除后重试)\n", p.Name)
			skipped++
			continue
		}

		if err := r.cfg.AddProfile(p); err != nil {
			PrintWarning("添加配置 '%s' 失败: %v", p.Name, err)
			continue
		}
		imported++
	}

	if err := r.cfg.Save(r.cfgPath); err != nil {
		PrintError("保存配置失败: %v", err)
		return
	}

	PrintSuccess("导入完成: %d 个配置已添加, %d 个已跳过", imported, skipped)
}

// cmdHistory 显示命令历史
func (r *REPL) cmdHistory(args []string) {
	commands := r.history.GetCommands()
	if len(commands) == 0 {
		PrintInfo("暂无命令历史")
		return
	}

	// 显示最近的 20 条命令
	start := 0
	if len(commands) > 20 {
		start = len(commands) - 20
	}

	fmt.Println()
	PrintInfo("最近 %d 条命令:", len(commands)-start)
	for i := start; i < len(commands); i++ {
		fmt.Printf("  %3d  %s\n", i+1, commands[i])
	}
	fmt.Println()
}

// cmdHelp 显示帮助
func (r *REPL) cmdHelp(args []string) {
	// 如果有参数，显示特定命令的详细帮助
	if len(args) > 0 {
		r.showCommandHelp(args[0])
		return
	}

	fmt.Println()
	fmt.Println("可用命令:")
	fmt.Println()

	// 配置管理
	fmt.Println("配置管理:")
	fmt.Println("  list, ls          列出所有配置")
	fmt.Println("  use, switch       切换当前会话配置")
	fmt.Println("  current, status   显示当前配置")
	fmt.Println("  default           设置默认配置")
	fmt.Println("  show              显示配置详情")
	fmt.Println("  add, new          添加配置（提示使用 setup）")
	fmt.Println("  edit              编辑配置")
	fmt.Println("  delete, rm        删除配置")
	fmt.Println("  copy, cp          复制配置")
	fmt.Println("  rename, mv        重命名配置")
	fmt.Println()

	// 测试与导入导出
	fmt.Println("测试与导入导出:")
	fmt.Println("  test              测试 API 连通性")
	fmt.Println("  export            导出配置到 stdout 或文件")
	fmt.Println("  import            从文件导入配置")
	fmt.Println()

	// 辅助命令
	fmt.Println("辅助命令:")
	fmt.Println("  history           显示命令历史")
	fmt.Println("  help, ?           显示帮助")
	fmt.Println("  clear, cls        清屏")
	fmt.Println("  exit, quit, q     退出")
	fmt.Println()

	// 启动
	fmt.Println("启动 Claude Code:")
	fmt.Println("  run [profile] [-- args...]  使用当前或指定配置启动")
	fmt.Println("  setup             运行配置向导")
	fmt.Println()
}

// showCommandHelp 显示特定命令的详细帮助
func (r *REPL) showCommandHelp(cmd string) {
	helpTexts := map[string]string{
		"list": `list, ls - 列出所有配置

用法: list

显示所有已配置的供应商，包括名称、Base URL、模型和状态。
状态标记:
  默认 - 默认配置
  当前 - 当前会话使用的配置`,
		"use": `use, switch - 切换当前会话配置

用法: use <name>

切换当前 REPL 会话使用的配置。
注意: 此命令只影响当前会话，不会修改默认配置。

示例:
  use moonshot    切换到 moonshot 配置`,
		"current": `current, status - 显示当前配置

用法: current

显示当前会话使用的配置详情，包括 Base URL、模型和 Token。`,
		"default": `default - 设置或显示默认配置

用法:
  default           显示当前默认配置
  default <name>    设置指定配置为默认

设置的默认配置会持久化到配置文件。

示例:
  default           显示当前默认配置
  default moonshot  将 moonshot 设为默认`,
		"show": `show - 显示配置详情

用法: show [name]

显示指定配置的详细信息。如果省略名称，显示当前配置。

示例:
  show           显示当前配置详情
  show moonshot  显示 moonshot 配置详情`,
		"add": `add, new - 添加配置

用法: add

启动交互式配置向导添加新的供应商配置。
建议直接使用 'setup' 命令。`,
		"edit": `edit - 编辑配置

用法: edit [name]

启动交互式向导编辑现有配置。如果省略名称，编辑当前配置。

示例:
  edit           编辑当前配置
  edit moonshot  编辑 moonshot 配置`,
		"delete": `delete, rm - 删除配置

用法: delete <name>

删除指定的配置。删除前会要求确认。

示例:
  delete moonshot  删除 moonshot 配置`,
		"copy": `copy, cp - 复制配置

用法: copy <source> <target>

复制现有配置到新名称。

示例:
  copy moonshot moonshot-backup  复制 moonshot 到 moonshot-backup`,
		"rename": `rename, mv - 重命名配置

用法: rename <old> <new>

重命名配置。会自动更新默认配置引用。

示例:
  rename moonshot kimi  将 moonshot 重命名为 kimi`,
		"test": `test - 测试 API 连通性

用法: test [name]

测试指定配置的 API 端点连通性。如果省略名称，测试当前配置。

示例:
  test           测试当前配置
  test moonshot  测试 moonshot 的 API 连通性`,
		"export": `export - 导出配置

用法: export [file]

导出配置到 JSON 格式。如果指定文件，保存到文件；否则输出到 stdout。

示例:
  export                    输出配置到屏幕
  export backup.json        保存配置到 backup.json`,
		"import": `import - 导入配置

用法: import <file>

从 JSON 文件导入配置。已存在的同名配置会被跳过。

示例:
  import backup.json  从 backup.json 导入配置`,
		"history": `history - 显示命令历史

用法: history

显示最近 20 条执行的命令。`,
		"run": `run - 启动 Claude Code

用法: run [profile] [-- args...]

使用指定或当前配置启动 Claude Code。

示例:
  run                    使用当前配置启动
  run moonshot           使用 moonshot 配置启动
  run -- --help          使用当前配置启动，传递 --help 给 claude
  run moonshot -- --help 使用 moonshot 配置启动，传递 --help`,
		"setup": `setup - 运行配置向导

用法: setup

启动交互式配置向导，添加新的供应商配置。`,
		"clear": `clear, cls - 清屏

用法: clear

清除终端屏幕。`,
		"exit": `exit, quit, q - 退出 REPL

用法: exit

退出 CC-Start REPL。`,
	}

	// 标准化命令名
	normalizedCmd := cmd
	aliases := map[string]string{
		"ls":     "list",
		"switch": "use",
		"status": "current",
		"new":    "add",
		"rm":     "delete",
		"cp":     "copy",
		"mv":     "rename",
		"?":      "help",
		"cls":    "clear",
		"quit":   "exit",
		"q":      "exit",
	}
	if n, ok := aliases[cmd]; ok {
		normalizedCmd = n
	}

	if help, ok := helpTexts[normalizedCmd]; ok {
		fmt.Println()
		fmt.Println(help)
		fmt.Println()
	} else {
		PrintError("未知命令: %s", cmd)
		PrintInfo("输入 'help' 查看所有可用命令")
	}
}

// cmdClear 清屏
func (r *REPL) cmdClear(args []string) {
	fmt.Print("\033[2J\033[H")
}

// cmdExit 退出
func (r *REPL) cmdExit(args []string) {
	fmt.Println("再见!")
	os.Exit(0)
}

// cmdRun 启动 Claude Code
// 用法: run [profile] [-- args...]
// - 无参数：使用当前配置启动
// - run <profile>：使用指定配置启动
// - run -- <args>：使用当前配置启动，传递参数给 claude
// - run <profile> -- <args>：使用指定配置启动，传递参数给 claude
func (r *REPL) cmdRun(args []string) {
	profileName := r.currentName
	launchArgs := args

	// 解析参数
	for i, arg := range args {
		if arg == "--" {
			// -- 后的参数传递给 claude
			launchArgs = args[i+1:]
			if i > 0 {
				// -- 前有参数，作为 profile 名称
				profileName = args[0]
			}
			break
		}
	}

	// 如果没有 -- 分隔符，检查第一个参数是否为 profile 名称
	if len(args) > 0 && len(launchArgs) == len(args) {
		// 尝试将第一个参数作为 profile 名称
		if _, err := r.cfg.GetProfile(args[0]); err == nil {
			profileName = args[0]
			launchArgs = args[1:]
		}
	}

	if profileName == "" {
		PrintError("请先选择配置: use <name>")
		PrintInfo("或指定配置名: run <profile>")
		return
	}

	profile, err := r.cfg.GetProfile(profileName)
	if err != nil {
		PrintError("配置无效: %v", err)
		return
	}

	// 保存历史，确保退出后历史不丢失
	r.history.Add("run " + strings.Join(args, " "))

	if err := launcher.Launch(profile, launchArgs); err != nil {
		PrintError("启动失败: %v", err)
	}
}

// cmdSetup 运行配置向导
func (r *REPL) cmdSetup(args []string) {
	m := setup.InitialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	result, err := p.Run()
	if err != nil {
		PrintError("启动 TUI 失败: %v", err)
		return
	}

	// 重新加载配置
	cfg, err := config.LoadConfig(r.cfgPath)
	if err != nil {
		PrintError("重新加载配置失败: %v", err)
		return
	}
	r.cfg = cfg

	if final, ok := result.(setup.Model); ok && final.Done() {
		PrintSuccess("配置 '%s' 已保存", final.GetName())
	}
}

// ExecuteCommand 执行命令
func (r *REPL) ExecuteCommand(cmd string, args []string) {
	switch cmd {
	// 配置查看
	case "list", "ls":
		r.cmdList(args)
	case "use", "switch":
		r.cmdUse(args)
	case "current", "status":
		r.cmdCurrent(args)
	case "default":
		r.cmdDefault(args)
	case "show":
		r.cmdShow(args)

	// 配置管理
	case "add", "new":
		r.cmdAdd(args)
	case "edit":
		r.cmdEdit(args)
	case "delete", "rm":
		r.cmdDelete(args)
	case "copy", "cp":
		r.cmdCopy(args)
	case "rename", "mv":
		r.cmdRename(args)

	// 测试与导入导出
	case "test":
		r.cmdTest(args)
	case "export":
		r.cmdExport(args)
	case "import":
		r.cmdImport(args)

	// 辅助命令
	case "history":
		r.cmdHistory(args)
	case "help", "?":
		r.cmdHelp(args)
	case "clear", "cls":
		r.cmdClear(args)
	case "exit", "quit", "q":
		r.cmdExit(args)

	// 启动
	case "run":
		r.cmdRun(args)
	case "setup":
		r.cmdSetup(args)

	default:
		PrintError("未知命令: %s", cmd)
		PrintInfo("输入 'help' 查看可用命令")
	}
}
