// internal/repl/repl.go
package repl

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/wujunwei/cc-start/internal/config"
)

// REPL 交互式 REPL
type REPL struct {
	cfg         *config.Config
	cfgPath     string
	completer   *Completer
	history     *History
	currentName string
}

// New 创建 REPL 实例
func New(cfgPath string) (*REPL, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	repl := &REPL{
		cfg:     cfg,
		cfgPath: cfgPath,
		history: NewHistory(),
	}

	repl.currentName = cfg.Default

	repl.completer = NewCompleter(func() []string {
		return repl.getProfileNames()
	})

	return repl, nil
}

// Run 启动 REPL
func (r *REPL) Run() {
	r.printWelcome()

	if len(r.cfg.Profiles) == 0 {
		PrintWarning("尚未配置任何供应商")
		fmt.Println("运行 'setup' 创建配置，或 'help' 查看帮助")
		fmt.Println()
	}

	p := prompt.New(
		r.executor,
		r.completer.Complete,
		prompt.OptionTitle("cc-start"),
		prompt.OptionPrefix(r.getPromptPrefix()),
		prompt.OptionLivePrefix(r.changeLivePrefix),
		prompt.OptionHistory(r.history.GetCommands()),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlL,
			Fn: func(buf *prompt.Buffer) {
				fmt.Print("\033[2J\033[H")
			},
		}),
	)

	p.Run()
}

func (r *REPL) printWelcome() {
	fmt.Println()
	fmt.Println("CC-Start REPL v1.0")
	fmt.Println("输入 'help' 查看可用命令，'exit' 退出。")
	fmt.Println()
}

func (r *REPL) getPromptPrefix() string {
	if r.currentName != "" {
		return fmt.Sprintf("cc-start [%s]> ", r.currentName)
	}
	return "cc-start> "
}

func (r *REPL) changeLivePrefix() (string, bool) {
	return r.getPromptPrefix(), true
}

func (r *REPL) executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	r.history.Add(in)

	parts := strings.Fields(in)
	cmd := parts[0]
	args := parts[1:]

	r.executeCommand(cmd, args)
}

func (r *REPL) getProfileNames() []string {
	names := make([]string, 0, len(r.cfg.Profiles))
	for _, p := range r.cfg.Profiles {
		names = append(names, p.Name)
	}
	return names
}

// executeCommand 执行命令（占位，将在 Task 3 实现）
func (r *REPL) executeCommand(cmd string, args []string) {
	switch cmd {
	case "help", "?":
		fmt.Println("命令处理器将在 Task 3 实现")
	case "exit", "quit", "q":
		fmt.Println("再见!")
	default:
		fmt.Printf("未知命令: %s\n", cmd)
	}
}
