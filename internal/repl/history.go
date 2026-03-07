// internal/repl/history.go
package repl

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxHistorySize = 1000
	historyFile    = "history"
)

// History 管理命令历史
type History struct {
	filePath string
	commands []string
}

// NewHistory 创建历史管理器
func NewHistory() *History {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".cc-start", historyFile)

	h := &History{
		filePath: filePath,
		commands: make([]string, 0),
	}
	h.load()
	return h
}

// load 从文件加载历史
func (h *History) load() {
	file, err := os.Open(h.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.commands = append(h.commands, line)
		}
	}
}

// Add 添加命令到历史
func (h *History) Add(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	// 避免重复的连续命令
	if len(h.commands) > 0 && h.commands[len(h.commands)-1] == cmd {
		return
	}

	h.commands = append(h.commands, cmd)

	// 超过限制时移除旧记录
	if len(h.commands) > maxHistorySize {
		h.commands = h.commands[1:]
	}

	h.save()
}

// save 保存历史到文件
func (h *History) save() {
	// 确保目录存在
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	file, err := os.Create(h.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	for _, cmd := range h.commands {
		file.WriteString(cmd + "\n")
	}
}

// GetCommands 获取历史命令列表
func (h *History) GetCommands() []string {
	return h.commands
}
