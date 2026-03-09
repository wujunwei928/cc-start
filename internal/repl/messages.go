// internal/repl/messages.go
package repl

// 消息类型定义
type (
	// CommandExecutedMsg 命令执行完成
	CommandExecutedMsg struct {
		Output string
		Err    error
	}

	// CommandSelectedMsg 从命令面板选择命令
	CommandSelectedMsg struct {
		Cmd  string
		Args []string
	}

	// ProfileChangedMsg 配置切换
	ProfileChangedMsg struct {
		Name string
	}

	// OutputClearedMsg 清空输出
	OutputClearedMsg struct{}
)
