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

type Manager struct {
	currentLang  string
	translations map[string]map[string]string
}

func NewManager() *Manager {
	m := &Manager{
		currentLang: LangZH,
	}
	m.loadTranslations()
	return m
}

func (m *Manager) SetLanguage(lang string) error {
	for _, supported := range supportedLanguages {
		if lang == supported {
			m.currentLang = lang
			return nil
		}
	}
	return fmt.Errorf("unsupported language: %s", lang)
}

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

func (m *Manager) TWithData(key string, data map[string]interface{}) string {
	return m.T(key)
}

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

func getZhTranslations() map[string]string {
	return map[string]string{}
}

func getEnTranslations() map[string]string {
	return map[string]string{}
}

func getJaTranslations() map[string]string {
	return map[string]string{}
}
