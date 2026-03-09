// internal/repl/autocomplete_test.go
package repl

import (
	"testing"

	"github.com/wujunwei928/cc-start/internal/i18n"
)

func TestAutocompleteNew(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	if ac == nil {
		t.Fatal("NewAutocomplete 返回 nil")
	}
	if ac.IsVisible() {
		t.Error("新建的 Autocomplete 不应该可见")
	}
	if ac.maxShow != 6 {
		t.Errorf("maxShow 应该是 6，实际是 %d", ac.maxShow)
	}
}

func TestAutocompleteShowHide(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	// 测试 Show
	ac.Show("/")
	if !ac.IsVisible() {
		t.Error("Show 后应该可见")
	}

	// 测试 Hide
	ac.Hide()
	if ac.IsVisible() {
		t.Error("Hide 后不应该可见")
	}
}
