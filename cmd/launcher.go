// cmd/launcher.go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/launcher"
)

var (
	launchModel   string
	launchBaseURL string
	launchToken   string
	launchEnv     []string
)

// runLaunchWithTool 使用指定工具名执行启动逻辑
// cmdName 用于在 os.Args 中定位参数位置
func runLaunchWithTool(toolName string, args []string, cmdName string) error {
	// 解析 profile 和工具参数
	var profileName string
	var toolArgs []string

	dashPos := findDashSeparator(os.Args)

	if dashPos != -1 {
		// 有 -- 分隔符
		toolArgs = os.Args[dashPos+1:]
		// 找命令之后、-- 之前的非 flag 参数作为 profile
		for i := dashPos - 1; i >= 0; i-- {
			if os.Args[i] == cmdName {
				for j := i + 1; j < dashPos; j++ {
					arg := os.Args[j]
					if !isFlag(arg) && arg != toolName && !isFlagValue(os.Args, j) {
						profileName = arg
						break
					}
				}
				break
			}
		}
	} else if len(args) > 0 {
		// 无 -- 分隔符，第一个非 flag 参数是 profile
		for _, arg := range args {
			if !isFlag(arg) && !isFlagValue(os.Args, findArgIndex(os.Args, arg)) {
				profileName = arg
				break
			}
		}
	}

	// 加载配置
	cfgPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 获取 profile（未指定时使用默认配置）
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w\n\n运行 'cc-start list' 查看可用配置", err)
	}

	// 解析环境变量
	envMap := make(map[string]string)
	for _, e := range launchEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// 构建启动配置
	launchCfg := &launcher.LaunchConfig{
		Tool:     toolName,
		Profile:  profile,
		Model:    launchModel,
		BaseURL:  launchBaseURL,
		Token:    launchToken,
		Env:      envMap,
		ToolArgs: toolArgs,
	}

	// 验证必要的配置
	if profile == nil && launchToken == "" {
		return fmt.Errorf("请指定 profile，运行 'cc-start list' 查看可用配置")
	}

	return launcher.LaunchWithTool(launchCfg)
}

// findArgIndex 查找参数在数组中的索引
func findArgIndex(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}

// isFlagValue 检查指定索引是否是某个 flag 的值
func isFlagValue(args []string, index int) bool {
	if index <= 0 || index >= len(args) {
		return false
	}
	return isFlag(args[index-1])
}
