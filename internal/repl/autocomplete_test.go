// internal/repl/autocomplete_test.go
package repl

import (
	"strings"
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

func TestAutocompleteFilter(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)

	// 测试 "/" 过滤 - 应该显示所有以 "/" 开头的命令
	ac.Show("/")
	filtered := ac.FilteredItems()
	if len(filtered) == 0 {
		t.Error("过滤 '/' 应该返回至少一个命令")
	}
	// 检查所有返回的命令都以 "/" 开头
	for _, item := range filtered {
		if !strings.HasPrefix(item.Cmd, "/") {
			t.Errorf("命令 %s 不以 '/' 开头", item.Cmd)
		}
	}

	// 测试 "/u" 过滤 - 应该只显示 /use 等
	ac.Filter("/u")
	filtered = ac.FilteredItems()
	if len(filtered) == 0 {
		t.Error("过滤 '/u' 应该返回至少一个命令")
	}
	for _, item := range filtered {
		if !strings.HasPrefix(item.Cmd, "/u") {
			t.Errorf("命令 %s 不以 '/u' 开头", item.Cmd)
		}
	}

	// 测试无匹配的情况
	ac.Filter("/xyz123notexist")
	filtered = ac.FilteredItems()
	if len(filtered) != 0 {
		t.Errorf("过滤不存在的命令应该返回空，实际返回 %d 条", len(filtered))
	}
}

func TestAutocompleteSelectUpAndDown(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)
	ac.Show("/")

	// 初始选中应该是 0
	if ac.SelectedIndex() != 0 {
		t.Errorf("初始选中应该是 0，实际是 %d", ac.SelectedIndex())
	}

	// 在第一项时向上不应该越界
	ac.SelectUp()
	if ac.SelectedIndex() != 0 {
		t.Errorf("在第一项时向上应该保持在 0，实际是 %d", ac.SelectedIndex())
	}

	// 向下选择
	ac.SelectDown()
	if ac.SelectedIndex() != 1 {
		t.Errorf("向下选择后应该是 1，实际是 %d", ac.SelectedIndex())
	}

	// 再向上回到第一项
	ac.SelectUp()
	if ac.SelectedIndex() != 0 {
		t.Errorf("向上选择后应该是 0，实际是 %d", ac.SelectedIndex())
	}
}

func TestAutocompleteSelectedCommand(t *testing.T) {
	i18nMgr := i18n.NewManager()
	styles := DefaultStyles()

	ac := NewAutocomplete(styles, i18nMgr)
	ac.Show("/")

	// 选中第一个命令
	cmd := ac.SelectedCommand()
	if cmd == "" {
		t.Error("SelectedCommand 不应该返回空")
	}
	if !strings.HasPrefix(cmd, "/") {
		t.Errorf("命令应该以 '/' 开头，实际是 %s", cmd)
	}

	// 隐藏后应该返回空
	ac.Hide()
	cmd = ac.SelectedCommand()
	if cmd != "" {
		t.Errorf("隐藏后 SelectedCommand 应该返回空，实际返回 %s", cmd)
	}
}
