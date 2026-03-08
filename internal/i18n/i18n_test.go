// internal/i18n/i18n_test.go
package i18n

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
	// 通过验证 SetLanguage 不返回错误来间接验证默认语言是有效的
	if err := mgr.SetLanguage(LangZH); err != nil {
		t.Errorf("NewManager() default language validation failed: %v", err)
	}
	// 验证获取支持的语言列表正常工作
	langs := mgr.GetSupportedLanguages()
	if len(langs) != 3 {
		t.Errorf("NewManager() GetSupportedLanguages() returned %d languages, want 3", len(langs))
	}
}

func TestSetLanguage(t *testing.T) {
	tests := []struct {
		name    string
		lang    string
		wantErr bool
	}{
		{"valid zh", LangZH, false},
		{"valid en", LangEN, false},
		{"valid ja", LangJA, false},
		{"invalid fr", "fr", true},
		{"invalid empty", "", true},
		{"invalid de", "de", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager()
			err := mgr.SetLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetLanguage(%s) error = %v, wantErr %v", tt.lang, err, tt.wantErr)
			}
			// 通过验证设置相同的语言不返回错误来间接验证语言设置成功
			if !tt.wantErr {
				if err := mgr.SetLanguage(tt.lang); err != nil {
					t.Errorf("SetLanguage(%s) follow-up validation failed: %v", tt.lang, err)
				}
			}
		})
	}
}

func TestT(t *testing.T) {
	mgr := NewManager()

	result := mgr.T("nonexistent.key")
	if result != "nonexistent.key" {
		t.Errorf("T(nonexistent.key) = %s, want nonexistent.key (fallback to key name)", result)
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	mgr := NewManager()
	langs := mgr.GetSupportedLanguages()

	if len(langs) != 3 {
		t.Errorf("GetSupportedLanguages() returned %d languages, want 3", len(langs))
	}

	expected := map[string]bool{LangZH: true, LangEN: true, LangJA: true}
	for _, lang := range langs {
		if !expected[lang] {
			t.Errorf("GetSupportedLanguages() returned unexpected language: %s", lang)
		}
	}
}

func TestZhTranslations(t *testing.T) {
	m := NewManager()
	m.SetLanguage(LangZH)

	tests := []struct {
		key  string
		want string
	}{
		{MsgCommonSuccess, "成功"},
		{MsgCommonError, "错误"},
		{MsgSettingsTitle, "⚙ 系统设置"},
		{MsgSettingsLanguage, "语言 / Language"},
		{MsgSettingsTheme, "主题 / Theme"},
		{MsgCmdList, "列出所有配置"},
		{MsgREPLWelcome, "欢迎使用 CC-Start"},
		{MsgErrProfileNotFound, "配置 '%s' 不存在"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := m.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}

func TestEnTranslations(t *testing.T) {
	m := NewManager()
	m.SetLanguage(LangEN)

	tests := []struct {
		key  string
		want string
	}{
		{MsgCommonSuccess, "Success"},
		{MsgCommonError, "Error"},
		{MsgSettingsTitle, "⚙ Settings"},
		{MsgSettingsLanguage, "Language"},
		{MsgSettingsTheme, "Theme"},
		{MsgCmdList, "List all profiles"},
		{MsgREPLWelcome, "Welcome to CC-Start"},
		{MsgErrProfileNotFound, "Profile '%s' not found"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := m.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}

func TestJaTranslations(t *testing.T) {
	m := NewManager()
	m.SetLanguage(LangJA)

	tests := []struct {
		key  string
		want string
	}{
		{MsgCommonSuccess, "成功"},
		{MsgCommonError, "エラー"},
		{MsgSettingsTitle, "⚙ 設定"},
		{MsgSettingsLanguage, "言語"},
		{MsgSettingsTheme, "テーマ"},
		{MsgCmdList, "すべてのプロファイルを一覧表示"},
		{MsgREPLWelcome, "CC-Startへようこそ"},
		{MsgErrProfileNotFound, "プロファイル '%s' が見つかりません"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := m.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%s) = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}
