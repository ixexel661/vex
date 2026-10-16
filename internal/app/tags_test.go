package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/theme"
)

func TestTypingStartTagAddsEndTag(t *testing.T) {
	m, _ := newModel(t, "")
	press(m, key("i"))
	press(m, typing("<b>")...)
	if got := m.editor.Document().Buffer.Text(); got != "<b></b>" {
		t.Errorf("text = %q, want %q", got, "<b></b>")
	}
	if got := m.editor.Cursor(); got != at(0, 3) {
		t.Errorf("cursor = %v, want it between the tags at {0 3}", got)
	}
}

func TestPastedStartTagStaysOpen(t *testing.T) {
	m, _ := newModel(t, "")
	press(m, key("i"))
	m.Update(tea.PasteMsg{Content: "<b>"})
	if got := m.editor.Document().Buffer.Text(); got != "<b>" {
		t.Errorf("text = %q, want %q", got, "<b>")
	}
}

func TestTypingInTagNameChangesBoth(t *testing.T) {
	m, _ := newModel(t, "<book>x</book>")
	press(m, key("i"))
	m.editor.MoveTo(at(0, 5))
	press(m, key("s"))
	if got := m.editor.Document().Buffer.Text(); got != "<books>x</books>" {
		t.Errorf("text = %q, want %q", got, "<books>x</books>")
	}
}

func TestTagPairIsMarkedInInsertMode(t *testing.T) {
	th := theme.Theme{Syntax: map[highlight.Kind]lipgloss.Style{
		highlight.LinkedTag: lipgloss.NewStyle().Transform(func(s string) string { return "[" + s + "]" }),
	}}
	m, _ := newModelWithTheme(t, "<book>x</book>", th)
	m.editor.MoveTo(at(0, 1))
	expect := func(when, want string) {
		t.Helper()
		if got := strings.Split(m.editor.View(), "\n")[0]; !strings.HasSuffix(got, want) {
			t.Errorf("%s: row 0 = %q, want %q", when, got, want)
		}
	}

	expect("normal mode", " <book>x</book>")
	press(m, key("i"))
	expect("insert mode", " <[book]>x</[book]>")
	press(m, esc)
	expect("back in normal mode", " <book>x</book>")
}
