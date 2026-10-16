package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/command"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/keymap"
	"github.com/ixexel661/vex/internal/mode"
	"github.com/ixexel661/vex/internal/search"
	"github.com/ixexel661/vex/internal/theme"
)

var (
	esc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	ctrlS = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
)

func key(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: []rune(text)[0], Text: text}
}

func newModel(t *testing.T, content string) (*Model, string) {
	t.Helper()
	return newModelWithTheme(t, content, theme.Theme{})
}

// newModelWithTheme keeps the document, the config file and the themes
// directory together in a temporary directory.
func newModelWithTheme(t *testing.T, content string, th theme.Theme) (*Model, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.xml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := document.Open(path, xml.Type())
	if err != nil {
		t.Fatal(err)
	}
	m := New(doc, setup(dir, th))
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return m, path
}

// setup keeps the config file and the themes directory in dir.
func setup(dir string, th theme.Theme) Setup {
	return Setup{
		Commands:   command.NewRegistry(command.Builtins()...),
		Keys:       keymap.Default(),
		Matchers:   []search.Matcher{search.Text{}, search.Fuzzy{}},
		Config:     config.Default(),
		ConfigPath: filepath.Join(dir, "config.toml"),
		Theme:      th,
		Themes:     theme.NewStore(filepath.Join(dir, "themes"), os.DirFS("../../themes")),
		Version:    "v1.2.3",
	}
}

// press sends the keys one by one and returns the command of the last one.
func press(m *Model, keys ...tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	for _, k := range keys {
		_, cmd = m.Update(k)
	}
	return cmd
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func fileContent(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestEditAndSave(t *testing.T) {
	m, path := newModel(t, "<a/>\n")
	press(m, key("i"), key("x"), esc, ctrlS)
	if got := fileContent(t, path); got != "x<a/>\n" {
		t.Errorf("file = %q, want %q", got, "x<a/>\n")
	}
	if m.mode != mode.Normal {
		t.Errorf("mode = %v, want NORMAL", m.mode)
	}
}

func TestLettersAreTextInInsertMode(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	if isQuit(press(m, key("i"), key("q"))) {
		t.Fatal("q quit in insert mode")
	}
	if got := m.editor.Document().Buffer.Text(); got != "q<a/>" {
		t.Errorf("text = %q, want %q", got, "q<a/>")
	}
}

func TestQuitWithoutChanges(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	if !isQuit(press(m, key("q"))) {
		t.Error("q did not quit")
	}
}

func TestQuitWithChangesAsks(t *testing.T) {
	tests := []struct {
		name     string
		answer   tea.KeyPressMsg
		wantQuit bool
		wantFile string
	}{
		{"save", key("y"), true, "x<a/>"},
		{"discard", key("n"), true, "<a/>"},
		{"cancel", esc, false, "<a/>"},
		{"other keys are ignored", key("q"), false, "<a/>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, path := newModel(t, "<a/>")
			if isQuit(press(m, key("i"), key("x"), esc, key("q"))) {
				t.Fatal("q quit without asking")
			}
			if m.dialog == nil {
				t.Fatal("q did not ask about unsaved changes")
			}
			if got := isQuit(press(m, tt.answer)); got != tt.wantQuit {
				t.Errorf("quit = %v, want %v", got, tt.wantQuit)
			}
			if got := fileContent(t, path); got != tt.wantFile {
				t.Errorf("file = %q, want %q", got, tt.wantFile)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	for _, src := range []string{"<a><b/></a>", "<a>\n\t<b/>\n</a>\n"} {
		m, _ := newModel(t, src)
		press(m, key("f"))
		if got := m.editor.Document().Buffer.Text(); got != "<a>\n  <b/>\n</a>\n" {
			t.Errorf("text of %q = %q", src, got)
		}
	}
}

func TestValidationReportsErrors(t *testing.T) {
	m, _ := newModel(t, "<a></b>")
	if m.validation.err == nil {
		t.Fatal("invalid document passed validation")
	}
	if !strings.Contains(m.messageLine(), "Invalid XML") {
		t.Errorf("message line = %q, want the validation error", m.messageLine())
	}
	if !strings.Contains(m.statusBar(), "✗ line 1") {
		t.Errorf("status bar = %q, want the error line", m.statusBar())
	}
}

func TestHints(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	if got, want := m.hints(), "i insert · ctrl+s save · f format · ctrl+f search · s settings · q quit"; got != want {
		t.Errorf("hints = %q, want %q", got, want)
	}
}

func TestDefaultBindingsUseExistingCommands(t *testing.T) {
	commands := command.NewRegistry(command.Builtins()...)
	keys := keymap.Default()
	for _, md := range []mode.Mode{mode.Normal, mode.Insert} {
		for _, b := range keys.Bindings(md) {
			if _, ok := commands.Lookup(b.Command); !ok {
				t.Errorf("%v key %q is bound to unknown command %q", md, b.Key, b.Command)
			}
		}
	}
}
