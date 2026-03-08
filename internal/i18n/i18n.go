// internal/i18n/i18n.go
package i18n

import (
	"fmt"
)

const (
	LangZH = "zh"
	LangEN = "en"
	LangJA = "ja"
)

var supportedLanguages = []string{LangZH, LangEN, LangJA}

// Manager 多语言管理器
type Manager struct {
	currentLang  string
	translations map[string]map[string]string
}

// NewManager 创建新的多语言管理器，默认语言为中文
func NewManager() *Manager {
	m := &Manager{
		currentLang: LangZH,
	}
	m.loadTranslations()
	return m
}

// SetLanguage 设置当前语言
func (m *Manager) SetLanguage(lang string) error {
	for _, supported := range supportedLanguages {
		if lang == supported {
			m.currentLang = lang
			return nil
		}
	}
	return fmt.Errorf("unsupported language: %s", lang)
}

// T 翻译指定键的文本
func (m *Manager) T(key string) string {
	if trans, ok := m.translations[m.currentLang]; ok {
		if msg, ok := trans[key]; ok {
			return msg
		}
	}
	if trans, ok := m.translations[LangEN]; ok {
		if msg, ok := trans[key]; ok {
			return msg
		}
	}
	return key
}

// TWithData 翻译指定键的文本，并替换变量
// TODO: 实现变量替换功能
func (m *Manager) TWithData(key string, data map[string]interface{}) string {
	text := m.T(key)
	// TODO: 实现变量替换
	return text
}

// GetSupportedLanguages 获取支持的语言列表
func (m *Manager) GetSupportedLanguages() []string {
	return supportedLanguages
}

func (m *Manager) loadTranslations() {
	m.translations = map[string]map[string]string{
		LangZH: getZhTranslations(),
		LangEN: getEnTranslations(),
		LangJA: getJaTranslations(),
	}
}


