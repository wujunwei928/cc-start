// internal/repl/integration_test.go
package repl

import (
	"testing"

	"github.com/wujunwei928/cc-start/internal/config"
	"github.com/wujunwei928/cc-start/internal/i18n"
	"github.com/wujunwei928/cc-start/internal/theme"
)

func TestLanguageSwitch(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := tmpDir + "/profiles.json"

	cfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "test", Token: "xxx"},
		},
		Settings: config.Settings{
			Language: "zh",
			Theme:    "default",
		},
	}
	cfg.Save(cfgPath)

	model, err := NewModel(cfgPath)
	if err != nil {
		t.Fatalf("NewModel() error: %v", err)
	}

	if model.I18n.T(i18n.MsgCommonSuccess) != "成功" {
		t.Errorf("Initial language not zh")
	}

	model.I18n.SetLanguage("en")
	if model.I18n.T(i18n.MsgCommonSuccess) != "Success" {
		t.Errorf("Language switch to en failed")
	}

	model.I18n.SetLanguage("ja")
	if model.I18n.T(i18n.MsgCommonSuccess) != "成功" {
		t.Errorf("Language switch to ja failed")
	}
}

func TestThemeSwitch(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := tmpDir + "/profiles.json"

	cfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "test", Token: "xxx"},
		},
		Settings: config.Settings{
			Language: "zh",
			Theme:    "default",
		},
	}
	cfg.Save(cfgPath)

	model, err := NewModel(cfgPath)
	if err != nil {
		t.Fatalf("NewModel() error: %v", err)
	}

	if model.Theme.Name != "default" {
		t.Errorf("Initial theme not default, got %s", model.Theme.Name)
	}

	oceanTheme, err := theme.GetTheme("ocean")
	if err != nil {
		t.Fatalf("GetTheme(ocean) error: %v", err)
	}

	model.Theme = oceanTheme
	model.Styles = NewStylesFromTheme(oceanTheme)

	if model.Theme.Name != "ocean" {
		t.Errorf("Theme not switched to ocean, got %s", model.Theme.Name)
	}

	if model.Theme.Colors.Primary != "#00bfff" {
		t.Errorf("Ocean theme primary color incorrect, got %s", model.Theme.Colors.Primary)
	}
}

func TestConfigPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := tmpDir + "/profiles.json"

	cfg := &config.Config{
		Profiles: []config.Profile{
			{Name: "test", Token: "xxx"},
		},
		Settings: config.Settings{
			Language: "zh",
			Theme:    "default",
		},
	}
	cfg.Save(cfgPath)

	cfg.UpdateSetting("language", "en")
	cfg.UpdateSetting("theme", "ocean")
	cfg.Save(cfgPath)

	loadedCfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loadedCfg.Settings.Language != "en" {
		t.Errorf("Language not saved, got %s", loadedCfg.Settings.Language)
	}

	if loadedCfg.Settings.Theme != "ocean" {
		t.Errorf("Theme not saved, got %s", loadedCfg.Settings.Theme)
	}
}
