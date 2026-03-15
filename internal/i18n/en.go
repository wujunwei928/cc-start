// internal/i18n/en.go
package i18n

// getEnTranslations 返回英文翻译
func getEnTranslations() map[string]string {
	return map[string]string{
		// Common
		MsgCommonSuccess: "Success",
		MsgCommonError:   "Error",
		MsgCommonInfo:    "Info",
		MsgCommonWarning: "Warning",

		// Settings Panel
		MsgSettingsTitle:    "⚙ Settings",
		MsgSettingsLanguage: "Language",
		MsgSettingsTheme:    "Theme",
		MsgSettingsHint:     "↑↓ Navigate  enter Confirm  esc Close",
		MsgSettingsCurrent:  "Current",

		// Command Palette
		MsgPaletteTitle:      "Command Palette",
		MsgPaletteSearchHint: "Type to search commands...",

		// REPL Interface
		MsgREPLInputPrompt: "Enter command...",
		MsgREPLWelcome:     "Welcome to CC-Start",
		MsgREPLHint:        "Type /help for available commands",

		// Command Descriptions
		MsgCmdList:    "List all profiles",
		MsgCmdUse:     "Switch current profile",
		MsgCmdSetup:   "Run setup wizard",
		MsgCmdEdit:    "Edit profile",
		MsgCmdDelete:  "Delete profile",
		MsgCmdCopy:    "Copy profile",
		MsgCmdRename:  "Rename profile",
		MsgCmdTest:    "Test API connectivity",
		MsgCmdExport:  "Export config to stdout or file",
		MsgCmdImport:  "Import config from file",
		MsgCmdRun:     "Launch with current or specified profile",
		MsgCmdHelp:    "Show help",
		MsgCmdExit:    "Exit",
		MsgCmdClear:   "Clear screen",
		MsgCmdHistory: "Show command history",
		MsgCmdDefault: "Set default profile",
		MsgCmdShow:    "Show profile details",
		MsgCmdCurrent: "Show current profile",

		// Error Messages
		MsgErrConfigLoad:      "Failed to load config: %s",
		MsgErrConfigSave:      "Failed to save config: %s",
		MsgErrInvalidLanguage: "Unsupported language: %s",
		MsgErrInvalidTheme:    "Unsupported theme: %s",
		MsgErrProfileNotFound: "Profile '%s' not found",
	}
}
