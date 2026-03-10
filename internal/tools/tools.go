// internal/tools/tools.go
package tools

import (
	"fmt"
	"sort"
)

// 参数类型常量
const (
	ParamToken   = "token"
	ParamBaseURL = "base_url"
)

// Tool 工具预设
type Tool struct {
	Name       string            // 工具名
	Executable string            // 可执行文件名
	EnvMap     map[string]string // 参数到环境变量的映射
}

// GetEnvName 获取指定参数对应的环境变量名
func (t *Tool) GetEnvName(param string) string {
	return t.EnvMap[param]
}

// 内置工具预设
var builtInTools = map[string]Tool{
	"claude": {
		Name:       "claude",
		Executable: "claude",
		EnvMap: map[string]string{
			ParamToken:   "ANTHROPIC_AUTH_TOKEN",
			ParamBaseURL: "ANTHROPIC_BASE_URL",
		},
	},
	"codex": {
		Name:       "codex",
		Executable: "codex",
		EnvMap: map[string]string{
			ParamToken:   "OPENAI_API_KEY",
			ParamBaseURL: "OPENAI_BASE_URL",
		},
	},
	"opencode": {
		Name:       "opencode",
		Executable: "opencode",
		EnvMap: map[string]string{
			ParamToken:   "OPENAI_API_KEY",
			ParamBaseURL: "OPENAI_BASE_URL",
		},
	},
}

// GetTool 获取工具预设
func GetTool(name string) (*Tool, error) {
	tool, ok := builtInTools[name]
	if !ok {
		return nil, fmt.Errorf("未知工具: %s\n可用工具: claude, codex, opencode", name)
	}
	return &tool, nil
}

// ListTools 返回所有可用工具名（按字母排序）
func ListTools() []string {
	names := make([]string, 0, len(builtInTools))
	for name := range builtInTools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
