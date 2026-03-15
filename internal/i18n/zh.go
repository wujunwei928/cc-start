// internal/i18n/zh.go
package i18n

// getZhTranslations 返回中文翻译
func getZhTranslations() map[string]string {
	return map[string]string{
		// 通用
		MsgCommonSuccess: "成功",
		MsgCommonError:   "错误",
		MsgCommonInfo:    "信息",
		MsgCommonWarning: "警告",

		// 设置面板
		MsgSettingsTitle:    "⚙ 系统设置",
		MsgSettingsLanguage: "语言 / Language",
		MsgSettingsTheme:    "主题 / Theme",
		MsgSettingsHint:     "↑↓ 导航  enter 确认  esc 关闭",
		MsgSettingsCurrent:  "当前",

		// 命令面板
		MsgPaletteTitle:      "命令面板",
		MsgPaletteSearchHint: "输入搜索命令...",

		// REPL 界面（placeholder 使用英文，避免 bubbletea 中文渲染 bug）
		MsgREPLInputPrompt: "Enter command...",
		MsgREPLWelcome:     "欢迎使用 CC-Start",
		MsgREPLHint:        "输入 /help 查看帮助",

		// 命令描述
		MsgCmdList:    "列出所有配置",
		MsgCmdUse:     "切换当前会话配置",
		MsgCmdSetup:   "运行配置向导",
		MsgCmdEdit:    "编辑配置",
		MsgCmdDelete:  "删除配置",
		MsgCmdCopy:    "复制配置",
		MsgCmdRename:  "重命名配置",
		MsgCmdTest:    "测试 API 连通性",
		MsgCmdExport:  "导出配置到 stdout 或文件",
		MsgCmdImport:  "从文件导入配置",
		MsgCmdRun:     "使用当前或指定配置启动",
		MsgCmdHelp:    "显示帮助",
		MsgCmdExit:    "退出",
		MsgCmdClear:   "清屏",
		MsgCmdHistory: "显示命令历史",
		MsgCmdDefault: "设置默认配置",
		MsgCmdShow:    "显示配置详情",
		MsgCmdCurrent: "显示当前配置",

		// 错误消息
		MsgErrConfigLoad:      "加载配置失败: %s",
		MsgErrConfigSave:      "保存配置失败: %s",
		MsgErrInvalidLanguage: "不支持的语言: %s",
		MsgErrInvalidTheme:    "不支持的主题: %s",
		MsgErrProfileNotFound: "配置 '%s' 不存在",
	}
}
