package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/command"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/keymap"
)

var (
	left  = tea.KeyPressMsg{Code: tea.KeyLeft}
	right = tea.KeyPressMsg{Code: tea.KeyRight}
)

// configPath is where newModel puts the config file next to the document.
func configPath(docPath string) string {
	return filepath.Join(filepath.Dir(docPath), "config.toml")
}

func TestSettingsPopup(t *testing.T) {
	m, _ := newModel(t, books)
	press(m, key("s"))
	if m.settings == nil {
		t.Fatal("s did not open the settings")
	}
	v := m.View()
	for _, want := range []string{
		"╭─ Settings ─", "▸ Theme", "‹ basic ›", "Tab width", "‹ 4 ›", "Indent", "‹ 2 spaces ›",
		"Empty elements", "‹ <a></a> ›", "Blank lines", "‹ remove ›", "Wrap tags", "‹ never ›", "config.toml │",
	} {
		if !strings.Contains(v.Content, want) {
			t.Errorf("view does not show %q:\n%s", want, v.Content)
		}
	}
	lines := strings.Split(v.Content, "\n")
	if len(lines) != 24 || !strings.Contains(lines[22], "NORMAL") || lines[23] != settingsHints {
		t.Errorf("want the popup over the editor, the status bar and the hints below it:\n%s", v.Content)
	}
	if v.Cursor != nil {
		t.Error("the cursor shows while the popup is open")
	}

	press(m, up)
	if !strings.Contains(m.View().Content, "▸ Wrap tags") {
		t.Error("up from the first setting did not select the last one")
	}
}

func TestSettingsApplyAtOnceAndCancel(t *testing.T) {
	m, path := newModel(t, "<a>\n\tb</a>")
	tab := func(width int) string {
		return "  2 " + strings.Repeat(" ", width) + "b"
	}

	press(m, key("s"), down, right)
	if m.Config().TabWidth != 8 || !strings.Contains(m.editor.View(), tab(8)) {
		t.Errorf("tab width = %d, want 8 at once:\n%s", m.Config().TabWidth, m.editor.View())
	}
	press(m, esc)
	if m.settings != nil {
		t.Fatal("esc did not close the settings")
	}
	if m.Config().TabWidth != 4 || !strings.Contains(m.editor.View(), tab(4)) {
		t.Errorf("tab width = %d, want 4 again after esc", m.Config().TabWidth)
	}
	if _, err := os.Stat(configPath(path)); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("esc wrote the config file: %v", err)
	}
}

func TestSettingsEnterSaves(t *testing.T) {
	m, path := newModel(t, books)
	press(m, key("s"), down, down, right, enter)
	if m.settings != nil {
		t.Fatal("enter did not close the settings")
	}
	if got := m.messageLine(); got != "Saved config.toml" {
		t.Errorf("message = %q", got)
	}
	want := config.Default()
	want.Format.Indent = "    "
	if got, err := config.Load(configPath(path)); err != nil || got != want {
		t.Errorf("saved config = %+v, %v, want %+v", got, err, want)
	}

	press(m, key("f"))
	if got := m.editor.Document().Buffer.Text(); !strings.HasPrefix(got, "<books>\n    <book>\n") {
		t.Errorf("format did not use the new indent:\n%s", got)
	}
}

func TestSettingsSwitchThemes(t *testing.T) {
	m, path := newModel(t, books)
	dir := filepath.Join(filepath.Dir(path), "themes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "loud.toml"), []byte("[ui]\nhint = { fg = \"red\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	renders := func(want lipgloss.Style) bool {
		return m.theme.Hint.Render("x") == want.Render("x")
	}

	press(m, key("s"), right)
	if m.Config().Theme != "loud" || !renders(lipgloss.NewStyle().Foreground(lipgloss.Red)) {
		t.Errorf("theme = %q, want loud with a red hint", m.Config().Theme)
	}
	press(m, right)
	if m.Config().Theme != "basic" || !renders(lipgloss.NewStyle().Foreground(lipgloss.BrightBlack)) {
		t.Errorf("theme = %q, want basic after wrapping around", m.Config().Theme)
	}
	press(m, esc)
	if m.Config().Theme != "basic" || !renders(lipgloss.NewStyle()) {
		t.Error("esc did not bring back the theme from before")
	}
}

func TestSettingsSaveErrorKeepsPopupOpen(t *testing.T) {
	m, path := newModel(t, books)
	if err := os.Mkdir(configPath(path), 0o755); err != nil {
		t.Fatal(err)
	}
	press(m, key("s"), enter)
	if m.settings == nil {
		t.Error("the popup closed although saving failed")
	}
	if !m.message.isError || !strings.HasPrefix(m.message.text, "cannot save the settings") {
		t.Errorf("message = %+v, want the error", m.message)
	}
}

func TestSettingsKeyTypesInInsertMode(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	press(m, key("i"), key("s"))
	if m.settings != nil {
		t.Error("s opened the settings in insert mode")
	}
	if got := m.editor.Document().Buffer.Text(); got != "s<a/>" {
		t.Errorf("text = %q, want %q", got, "s<a/>")
	}
}

func TestWarningShowsAtStart(t *testing.T) {
	doc, err := document.Open(filepath.Join(t.TempDir(), "new.xml"), xml.Type())
	if err != nil {
		t.Fatal(err)
	}
	m := New(doc, Setup{
		Commands: command.NewRegistry(),
		Keys:     keymap.New(),
		Warning:  errors.Join(errors.New("config.toml: unknown setting font"), errors.New(`theme "dark" not found, using basic`)),
	})
	if got, want := m.messageLine(), `config.toml: unknown setting font; theme "dark" not found, using basic`; got != want {
		t.Errorf("message line = %q, want %q", got, want)
	}
}

func TestCycle(t *testing.T) {
	choices := []int{2, 4, 8}
	tests := []struct{ current, step, want int }{
		{4, 1, 8},
		{8, 1, 2},
		{2, -1, 8},
		{3, 1, 2},
		{3, -1, 8},
	}
	for _, tt := range tests {
		if got := cycle(choices, tt.current, tt.step); got != tt.want {
			t.Errorf("cycle(%v, %d, %d) = %d, want %d", choices, tt.current, tt.step, got, tt.want)
		}
	}
}
