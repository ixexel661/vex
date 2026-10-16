package editor

import (
	"testing"

	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/theme"
)

// newIndentEditor returns an editor that indents with indent and has a tab
// width of 4. The cursor is at the "|" in text.
func newIndentEditor(t *testing.T, indent config.Indent, text string) *Editor {
	t.Helper()
	e := newEditorAt(t, filetype.Plain(), text)
	cfg := config.Default()
	cfg.Format.Indent = indent
	e.Configure(cfg, theme.Theme{})
	return e
}

type indentCase struct {
	name   string
	indent config.Indent
	text   string
	want   string
}

func runIndentCases(t *testing.T, edit func(*Editor), cases []indentCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			e := newIndentEditor(t, tt.indent, tt.text)
			edit(e)
			if got := withCursor(e); got != tt.want {
				t.Errorf("text = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIndent(t *testing.T) {
	runIndentCases(t, (*Editor).Indent, []indentCase{
		{"at the line start", "  ", "|<a>", "  |<a>"},
		{"to the next stop", "  ", "x|", "x |"},
		{"from a stop", "  ", "  |<a>", "    |<a>"},
		{"after an odd column", "  ", "   |<a>", "    |<a>"},
		{"after a tab", "  ", "\t|<a>", "\t  |<a>"},
		{"with 4 spaces", "    ", "x|", "x   |"},
		{"with tabs", "\t", "  |<a>", "  \t|<a>"},
	})
}

func TestOutdent(t *testing.T) {
	runIndentCases(t, (*Editor).Outdent, []indentCase{
		{"one level", "  ", "    <a|>", "  <a|>"},
		{"back to the previous stop", "  ", "   <a|>", "  <a|>"},
		{"less than a level", "  ", " <a|>", "<a|>"},
		{"cursor at the end of the indentation", "  ", "    |<a>", "  |<a>"},
		{"cursor in the removed part", "  ", "   | <a>", "  |<a>"},
		{"cursor at the line start", "  ", "|    <a>", "|  <a>"},
		{"a tab goes as a whole", "  ", "\t\t<a|>", "\t<a|>"},
		{"with tabs", "\t", "      <a|>", "    <a|>"},
	})
}

func TestOutdentWithoutIndentationChangesNothing(t *testing.T) {
	e := newIndentEditor(t, "  ", "<a|>")
	e.Outdent()
	if got := withCursor(e); got != "<a|>" || e.Document().Modified() {
		t.Errorf("text = %q, modified = %v, want it unchanged", got, e.Document().Modified())
	}
}

func TestBackspaceInIndentation(t *testing.T) {
	runIndentCases(t, (*Editor).DeleteBackward, []indentCase{
		{"one level", "  ", "    |<a>", "  |<a>"},
		{"back to the previous stop", "  ", "   |<a>", "  |<a>"},
		{"in the middle of the indentation", "  ", "    |  <a>", "  |  <a>"},
		{"spaces after a tab", "  ", "\t |<a>", "\t|<a>"},
		{"stops at a tab", "        ", "\t |<a>", "\t|<a>"},
		{"a tab alone", "  ", "\t|<a>", "|<a>"},
		{"after text", "  ", "  x   |", "  x  |"},
		{"at the line start", "  ", "a\n|  <b>", "a|  <b>"},
		{"with tabs", "\t", "        |<a>", "    |<a>"},
	})
}
