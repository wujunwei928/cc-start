// cmd/launch.go
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/launcher"
	"github.com/wujunwei928/cc-start/internal/tools"
)

var (
	launchModel   string
	launchBaseURL string
	launchToken   string
	launchEnv     []string
)

// launchCmd 启动 AI 工具命令
var launchCmd = &cobra.Command{
	Use:   "launch <tool> [profile] [flags] [-- tool-args]",
	Short: "启动 AI 编程助手",
	Long: `使用指定的 AI 工具和配置启动编程助手。

工具:
  claude    Anthropic Claude Code CLI
  codex     OpenAI Codex CLI
  opencode  OpenCode AI 编程助手

示例:
  cc-start launch claude                      使用默认配置启动 claude
  cc-start launch claude moonshot             使用 moonshot 配置
  cc-start launch codex -m gpt-4 -t sk-xxx    指定模型和令牌
  cc-start launch claude moonshot -e DEBUG=true -- --help`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLaunch,
}

func init() {
	rootCmd.AddCommand(launchCmd)

	launchCmd.Flags().StringVarP(&launchModel, "model", "m", "", "模型名称")
	launchCmd.Flags().StringVarP(&launchBaseURL, "base-url", "b", "", "API 基础地址")
	launchCmd.Flags().StringVarP(&launchToken, "token", "t", "", "认证令牌")
	launchCmd.Flags().StringArrayVarP(&launchEnv, "env", "e", nil, "环境变量 (格式: KEY=VALUE)")
}

func runLaunch(cmd *cobra.Command, args []string) error {
	// 第一个参数是工具名
	toolName := args[0]

	// 验证工具名
	if _, err := tools.GetTool(toolName); err != nil {
		return err
	}

	// 解析 profile 和工具参数
	var profileName string
	var toolArgs []string

	remainingArgs := args[1:]
	dashPos := findDashSeparator(os.Args)

	if dashPos != -1 {
		// 有 -- 分隔符
		toolArgs = os.Args[dashPos+1:]
		// 找 launch 之后、-- 之前的非 flag 参数作为 profile
		for i := dashPos - 1; i >= 0; i-- {
			if os.Args[i] == "launch" {
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
	} else if len(remainingArgs) > 0 {
		// 无 -- 分隔符，第一个非 flag 参数是 profile
		for _, arg := range remainingArgs {
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

	// 获取 profile（如果指定）
	var profile *config.Profile
	if profileName != "" {
		profile, err = cfg.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("获取配置失败: %w\n\n运行 'cc-start list' 查看可用配置", err)
		}
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
		return fmt.Errorf("请通过 -t 指定令牌或指定 profile")
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
