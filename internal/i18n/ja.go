// internal/i18n/ja.go
package i18n

// getJaTranslations 返回日文翻译
func getJaTranslations() map[string]string {
	return map[string]string{
		// 共通
		MsgCommonSuccess: "成功",
		MsgCommonError:   "エラー",
		MsgCommonInfo:    "情報",
		MsgCommonWarning: "警告",

		// 設定パネル
		MsgSettingsTitle:    "⚙ 設定",
		MsgSettingsLanguage: "言語",
		MsgSettingsTheme:    "テーマ",
		MsgSettingsHint:     "↑↓ 移動  enter 確定  esc 閉じる",
		MsgSettingsCurrent:  "現在",

		// コマンドパレット
		MsgPaletteTitle:      "コマンドパレット",
		MsgPaletteSearchHint: "コマンドを検索...",

		// REPL インターフェース
		MsgREPLInputPrompt: "コマンドを入力...",
		MsgREPLWelcome:     "CC-Startへようこそ",
		MsgREPLHint:        "/help でヘルプを表示",

		// コマンド説明
		MsgCmdList:    "すべてのプロファイルを一覧表示",
		MsgCmdUse:     "現在のプロファイルを切り替え",
		MsgCmdSetup:   "セットアップウィザードを実行",
		MsgCmdEdit:    "プロファイルを編集",
		MsgCmdDelete:  "プロファイルを削除",
		MsgCmdCopy:    "プロファイルをコピー",
		MsgCmdRename:  "プロファイル名を変更",
		MsgCmdTest:    "API接続をテスト",
		MsgCmdExport:  "設定をstdoutまたはファイルにエクスポート",
		MsgCmdImport:  "ファイルから設定をインポート",
		MsgCmdRun:     "現在または指定されたプロファイルで起動",
		MsgCmdHelp:    "ヘルプを表示",
		MsgCmdExit:    "終了",
		MsgCmdClear:   "画面をクリア",
		MsgCmdHistory: "コマンド履歴を表示",
		MsgCmdDefault: "デフォルトプロファイルを設定",
		MsgCmdShow:    "プロファイル詳細を表示",
		MsgCmdCurrent: "現在のプロファイルを表示",

		// エラーメッセージ
		MsgErrConfigLoad:      "設定の読み込みに失敗: %s",
		MsgErrConfigSave:      "設定の保存に失敗: %s",
		MsgErrInvalidLanguage: "サポートされていない言語: %s",
		MsgErrInvalidTheme:    "サポートされていないテーマ: %s",
		MsgErrProfileNotFound: "プロファイル '%s' が見つかりません",
	}
}
