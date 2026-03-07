// internal/repl/completer.go
package repl

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

// CommandDef 命令定义
type CommandDef struct {
	Name        string
	Aliases     []string
	Description string
	ArgsHint    string // 参数提示
	SubCommands []CommandDef
}

// Completer 补全器
type Completer struct {
	commands    []CommandDef
	getProfiles func() []string
}

// NewCompleter 创建补全器
func NewCompleter(getProfiles func() []string) *Completer {
	c := &Completer{
		getProfiles: getProfiles,
	}
	c.commands = c.getCommandDefs()
	return c
}

func (c *Completer) getCommandDefs() []CommandDef {
	return []CommandDef{
		{Name: "/list", Aliases: []string{"/ls"}, Description: "列出所有配置", ArgsHint: ""},
		{Name: "/use", Aliases: []string{"/switch"}, Description: "切换当前配置", ArgsHint: "<name>"},
		{Name: "/current", Aliases: []string{"/status"}, Description: "显示当前配置", ArgsHint: ""},
		{Name: "/default", Description: "设置默认配置", ArgsHint: "<name>"},
		{Name: "/show", Description: "显示配置详情", ArgsHint: "<name>"},
		{Name: "/add", Aliases: []string{"/new"}, Description: "添加新配置", ArgsHint: ""},
		{Name: "/edit", Description: "编辑配置", ArgsHint: "<name>"},
		{Name: "/delete", Aliases: []string{"/rm"}, Description: "删除配置", ArgsHint: "<name>"},
		{Name: "/copy", Aliases: []string{"/cp"}, Description: "复制配置", ArgsHint: "<from> <to>"},
		{Name: "/rename", Aliases: []string{"/mv"}, Description: "重命名配置", ArgsHint: "<old> <new>"},
		{Name: "/test", Description: "测试 API 连通性", ArgsHint: "<name>"},
		{Name: "/export", Description: "导出配置", ArgsHint: "[file]"},
		{Name: "/import", Description: "导入配置", ArgsHint: "<file>"},
		{Name: "/history", Description: "显示命令历史", ArgsHint: ""},
		{Name: "/help", Aliases: []string{"/?", "/h"}, Description: "显示帮助", ArgsHint: "[cmd]"},
		{Name: "/clear", Aliases: []string{"/cls"}, Description: "清屏", ArgsHint: ""},
		{Name: "/exit", Aliases: []string{"/quit", "/q"}, Description: "退出 REPL", ArgsHint: ""},
		{Name: "/run", Description: "启动 Claude Code", ArgsHint: "[profile] [-- args]"},
		{Name: "/setup", Description: "运行配置向导", ArgsHint: ""},
	}
}

// Complete 执行补全
func (c *Completer) Complete(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	words := strings.Fields(text)

	// 空输入或正在输入第一个命令
	if len(words) == 0 || (len(words) == 1 && !strings.HasSuffix(text, " ")) {
		return c.completeCommand(words)
	}

	// 获取命令名
	cmdName := words[0]

	// 需要配置名补全的命令
	needsProfile := map[string]bool{
		"/use": true, "/switch": true, "/show": true, "/edit": true,
		"/delete": true, "/rm": true, "/test": true, "/default": true,
		"/copy": true, "/cp": true, "/rename": true, "/mv": true, "/run": true,
	}

	if needsProfile[cmdName] {
		return c.completeProfile(words, d)
	}

	return []prompt.Suggest{}
}

// completeCommand 补全命令名（支持模糊过滤）
func (c *Completer) completeCommand(words []string) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0, len(c.commands)*2)

	for _, cmd := range c.commands {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        cmd.Name,
			Description: cmd.Description,
		})
		for _, alias := range cmd.Aliases {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        alias,
				Description: cmd.Description + " (alias)",
			})
		}
	}

	// 如果有输入，使用模糊过滤
	if len(words) > 0 {
		return prompt.FilterFuzzy(suggestions, words[0], true)
	}

	return suggestions
}

// completeProfile 补全配置名（支持模糊过滤）
func (c *Completer) completeProfile(words []string, d prompt.Document) []prompt.Suggest {
	if c.getProfiles == nil {
		return []prompt.Suggest{}
	}

	profiles := c.getProfiles()
	suggestions := make([]prompt.Suggest, 0, len(profiles))

	for _, p := range profiles {
		suggestions = append(suggestions, prompt.Suggest{
			Text: p,
		})
	}

	// 获取当前正在输入的参数，进行模糊过滤
	if len(words) >= 2 {
		currentArg := d.GetWordBeforeCursor()
		if currentArg != "" {
			return prompt.FilterFuzzy(suggestions, currentArg, true)
		}
	}

	return suggestions
}
