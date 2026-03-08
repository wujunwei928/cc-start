// internal/repl/integration_test.go
package repl

import (
	"os"
	"testing"

	"github.com/wujunwei/cc-start/internal/config"
	"github.com/wujunwei/cc-start/internal/i18n"
	"github.com/wujunwei/cc-start/internal/theme"
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

	if model.i18n.T(i18n.MsgCommonSuccess) != "成功" {
		t.Errorf("Initial language not zh")
	}

	model.i18n.SetLanguage("en")
	if model.i18n.T(i18n.MsgCommonSuccess) != "Success" {
		t.Errorf("Language switch to en failed")
	}

	model.i18n.SetLanguage("ja")
	if model.i18n.T(i18n.MsgCommonSuccess) != "成功" {
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

	if model.theme.Name != "default" {
		t.Errorf("Initial theme not default, got %s", model.theme.Name)
	}

	oceanTheme, err := theme.GetTheme("ocean")
	if err != nil {
		t.Fatalf("GetTheme(ocean) error: %v", err)
	}

	model.theme = oceanTheme
	model.styles = *NewStylesFromTheme(oceanTheme)

	if model.theme.Name != "ocean" {
		t.Errorf("Theme not switched to ocean, got %s", model.theme.Name)
	}

	if model.theme.Colors.Primary != "#00bfff" {
		t.Errorf("Ocean theme primary color incorrect, got %s", model.theme.Colors.Primary)
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
