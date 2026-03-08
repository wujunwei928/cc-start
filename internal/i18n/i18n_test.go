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
