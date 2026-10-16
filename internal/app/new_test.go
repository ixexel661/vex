package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/mode"
	"github.com/ixexel661/vex/internal/theme"
)

const template = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"

// newUnnamedModel returns the model of "vex new" and a directory to save in.
func newUnnamedModel(t *testing.T) (*Model, string) {
	t.Helper()
	dir := t.TempDir()
	s := setup(dir, theme.Theme{})
	s.NewFile = true
	m := New(document.New("", xml.Type()), s)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return m, dir
}

func paste(m *Model, text string) {
	m.Update(tea.PasteMsg{Content: text})
}

func TestNewFileStartsInInsertMode(t *testing.T) {
	m, _ := newUnnamedModel(t)
	if m.mode != mode.Insert {
		t.Errorf("mode = %v, want INSERT", m.mode)
	}
	if got := m.editor.Cursor(); got != (buffer.Position{Row: 1}) {
		t.Errorf("cursor = %v, want the line after the template", got)
	}
	if got := m.statusBar(); !strings.Contains(got, " untitled ") {
		t.Errorf("status bar = %q, want the name untitled", got)
	}
	if got := m.editor.View(); !strings.HasPrefix(got, "  1 <?xml") {
		t.Errorf("view = %q, want the template from its first line on", got)
	}
}

func TestSaveAsAsksForAName(t *testing.T) {
	m, dir := newUnnamedModel(t)
	press(m, key("x"), ctrlS)
	if m.prompt == nil {
		t.Fatal("ctrl+s did not ask for a name")
	}
	paste(m, filepath.Join(dir, "a")+"\nsecond line")
	if got, want := m.messageLine(), "Save as: "+filepath.Join(dir, "a")+"  "+promptHints; got != want {
		t.Errorf("message line = %q, want %q", got, want)
	}
	if c := m.View().Cursor; c.X != len("Save as: "+filepath.Join(dir, "a")) || c.Y != 23 {
		t.Errorf("cursor = %v, want it after the name in the message line", c.Position)
	}

	press(m, enter)
	if got := fileContent(t, filepath.Join(dir, "a.xml")); got != template+"x" {
		t.Errorf("file = %q, want %q", got, template+"x")
	}
	if got := m.messageLine(); got != "Saved a.xml" {
		t.Errorf("message = %q", got)
	}

	press(m, key("y"), ctrlS)
	if m.prompt != nil {
		t.Error("ctrl+s asked for a name again")
	}
	if got := fileContent(t, filepath.Join(dir, "a.xml")); got != template+"xy" {
		t.Errorf("file = %q, want %q", got, template+"xy")
	}
}

func TestSaveAsPromptKeys(t *testing.T) {
	m, dir := newUnnamedModel(t)
	press(m, ctrlS, enter)
	if m.prompt == nil {
		t.Fatal("enter without a name closed the prompt")
	}
	if isQuit(press(m, key("q"))) || m.prompt.text != "q" {
		t.Errorf("q in the prompt quit or was not typed, text = %q", m.prompt.text)
	}
	press(m, key("ä"), backspace)
	if m.prompt.text != "q" {
		t.Errorf("text after backspace = %q, want q", m.prompt.text)
	}
	press(m, esc)
	if m.prompt != nil {
		t.Fatal("esc did not close the prompt")
	}
	if m.mode != mode.Insert {
		t.Errorf("mode = %v, want INSERT as before", m.mode)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("esc wrote files: %v", entries)
	}
}

func TestSaveAsAsksBeforeOverwriting(t *testing.T) {
	m, dir := newUnnamedModel(t)
	path := filepath.Join(dir, "a.xml")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	press(m, ctrlS)
	paste(m, path)
	press(m, enter)
	if m.dialog == nil || !strings.Contains(m.messageLine(), "a.xml exists already. Overwrite it?") {
		t.Fatalf("no question about the existing file, message line = %q", m.messageLine())
	}
	press(m, esc)
	if got := fileContent(t, path); got != "old" || m.editor.Document().Path != "" {
		t.Errorf("esc overwrote the file (%q) or named the document %q", got, m.editor.Document().Path)
	}

	press(m, ctrlS)
	paste(m, path)
	press(m, enter, key("y"))
	if got := fileContent(t, path); got != template {
		t.Errorf("file = %q, want it overwritten", got)
	}
}

func TestQuitSavesUnnamedDocument(t *testing.T) {
	m, dir := newUnnamedModel(t)
	press(m, key("x"), esc, key("q"))
	if !strings.Contains(m.messageLine(), "Save changes to untitled?") {
		t.Fatalf("message line = %q, want the question about unsaved changes", m.messageLine())
	}
	press(m, key("y"))
	if m.prompt == nil {
		t.Fatal("saving did not ask for a name")
	}
	paste(m, filepath.Join(dir, "a"))
	if !isQuit(press(m, enter)) {
		t.Error("vex did not quit after saving")
	}
	if got := fileContent(t, filepath.Join(dir, "a.xml")); got != template+"x" {
		t.Errorf("file = %q", got)
	}
}

func TestQuitUnchangedUnnamedDocument(t *testing.T) {
	m, dir := newUnnamedModel(t)
	if !isQuit(press(m, esc, key("q"))) {
		t.Error("q did not quit a new document without changes")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("quitting wrote files: %v", entries)
	}
}
