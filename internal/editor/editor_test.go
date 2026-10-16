package editor

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/theme"
)

// newEditor returns an unstyled editor, so its view is plain text.
func newEditor(t *testing.T, text string, width, height int) *Editor {
	t.Helper()
	doc := &document.Document{Path: "test.txt", Type: filetype.Plain(), Buffer: buffer.New(text)}
	e := New(doc, config.Default(), theme.Theme{})
	e.SetSize(width, height)
	return e
}

// newEditorAt returns an unstyled editor with the cursor at the "|" in text.
func newEditorAt(t *testing.T, fileType filetype.Type, text string) *Editor {
	t.Helper()
	before, after, _ := strings.Cut(text, "|")
	doc := &document.Document{Path: "test.xml", Type: fileType, Buffer: buffer.New(before + after)}
	e := New(doc, config.Default(), theme.Theme{})
	e.SetSize(80, 1)
	lines := strings.Split(before, "\n")
	e.MoveTo(buffer.Position{Row: len(lines) - 1, Col: utf8.RuneCountInString(lines[len(lines)-1])})
	return e
}

// withCursor returns the text of e with the cursor as "|".
func withCursor(e *Editor) string {
	lines := strings.Split(e.Document().Buffer.Text(), "\n")
	c := e.Cursor()
	line := []rune(lines[c.Row])
	lines[c.Row] = string(line[:c.Col]) + "|" + string(line[c.Col:])
	return strings.Join(lines, "\n")
}

func TestNewlineKeepsIndentation(t *testing.T) {
	e := newEditor(t, "  <a>", 40, 10)
	e.MoveLineEnd()
	e.Newline()
	if got := e.Document().Buffer.Text(); got != "  <a>\n  " {
		t.Errorf("text = %q, want %q", got, "  <a>\n  ")
	}
	if got := e.Cursor(); got != (buffer.Position{Row: 1, Col: 2}) {
		t.Errorf("cursor = %v, want {1 2}", got)
	}
}

func TestDeleteBackwardJoinsLines(t *testing.T) {
	e := newEditor(t, "ab\ncd", 40, 10)
	e.MoveDown()
	e.DeleteBackward()
	if got := e.Document().Buffer.Text(); got != "abcd" {
		t.Errorf("text = %q, want %q", got, "abcd")
	}
	if got := e.Cursor(); got != (buffer.Position{Row: 0, Col: 2}) {
		t.Errorf("cursor = %v, want {0 2}", got)
	}
}

func TestVerticalMovementKeepsColumn(t *testing.T) {
	e := newEditor(t, "long line\nx\nlong line", 40, 10)
	e.MoveLineEnd()
	e.MoveDown()
	if got := e.Cursor(); got != (buffer.Position{Row: 1, Col: 1}) {
		t.Errorf("cursor on short line = %v, want {1 1}", got)
	}
	e.MoveDown()
	if got := e.Cursor(); got != (buffer.Position{Row: 2, Col: 9}) {
		t.Errorf("cursor on long line = %v, want {2 9}", got)
	}
}

func TestScrollingFollowsCursor(t *testing.T) {
	e := newEditor(t, strings.Repeat("line\n", 99), 40, 10)
	e.MoveBottom()
	if _, y := e.CursorPosition(); y != 9 {
		t.Errorf("cursor row on screen = %d, want 9", y)
	}
	lines := strings.Split(e.View(), "\n")
	if got := strings.TrimSpace(lines[9]); got != "100" {
		t.Errorf("last visible line = %q, want line number 100", got)
	}
}

func TestGrowingWindowShowsTextAbove(t *testing.T) {
	e := newEditor(t, "a\nb\nc", 40, 1)
	e.MoveBottom()
	e.SetSize(40, 10)
	if got := strings.Split(e.View(), "\n")[0]; got != "  1 a" {
		t.Errorf("first visible line = %q, want line 1", got)
	}
}

func TestTabsExpandToTabStops(t *testing.T) {
	e := newEditor(t, "a\tb", 40, 1)
	if got := e.View(); got != "  1 a   b" {
		t.Errorf("View() = %q, want %q", got, "  1 a   b")
	}
	e.MoveLineEnd()
	if x, _ := e.CursorPosition(); x != 4+5 {
		t.Errorf("cursor x = %d, want %d", x, 4+5)
	}
}

func TestConfigureChangesTabsAndStyles(t *testing.T) {
	e := newEditor(t, "a\tb", 40, 1)
	cfg := config.Default()
	cfg.TabWidth = 8
	dots := lipgloss.NewStyle().Transform(func(s string) string { return strings.ReplaceAll(s, " ", ".") })
	e.Configure(cfg, theme.Theme{CurrentLineNumber: dots})
	if got, want := e.View(), "..1.a       b"; got != want {
		t.Errorf("View() = %q, want %q", got, want)
	}
}

func TestHorizontalScrolling(t *testing.T) {
	e := newEditor(t, strings.Repeat("x", 30)+"END", 14, 1)
	e.MoveLineEnd()
	x, _ := e.CursorPosition()
	if x >= 14 {
		t.Errorf("cursor x = %d, outside the editor width 14", x)
	}
	if got := e.View(); !strings.HasSuffix(got, "END") {
		t.Errorf("View() = %q, want the end of the line to be visible", got)
	}
}

func TestSetTextKeepsCursorInBounds(t *testing.T) {
	e := newEditor(t, "a\nb\nc", 40, 10)
	e.MoveBottom()
	e.SetText("x")
	if got := e.Cursor(); got != (buffer.Position{Row: 0, Col: 1}) {
		t.Errorf("cursor = %v, want {0 1}", got)
	}
}

func TestMoveTo(t *testing.T) {
	e := newEditor(t, "abc\nd\nabcd", 40, 10)
	e.MoveTo(buffer.Position{Row: 9, Col: 9})
	if got := e.Cursor(); got != (buffer.Position{Row: 2, Col: 4}) {
		t.Errorf("cursor = %v, want it clamped to {2 4}", got)
	}
	e.MoveTo(buffer.Position{Row: 0, Col: 3})
	e.MoveDown()
	e.MoveDown()
	if got := e.Cursor(); got != (buffer.Position{Row: 2, Col: 3}) {
		t.Errorf("cursor = %v, want {2 3}, the column MoveTo went to", got)
	}
}

// tagAll highlights every line as a tag.
type tagAll struct{}

func (tagAll) Highlight(line []rune, state highlight.State) ([]highlight.Span, highlight.State) {
	return []highlight.Span{{Start: 0, End: len(line), Kind: highlight.Tag}}, state
}

func TestMarksOverrideSyntax(t *testing.T) {
	// The styles change the text instead of its color, so the view stays plain.
	th := theme.Theme{Syntax: map[highlight.Kind]lipgloss.Style{
		highlight.Tag:   lipgloss.NewStyle().Transform(strings.ToUpper),
		highlight.Match: lipgloss.NewStyle().Transform(func(s string) string { return "[" + s + "]" }),
	}}
	doc := &document.Document{Path: "test.xml", Type: filetype.Type{Highlighter: tagAll{}}, Buffer: buffer.New("book")}
	e := New(doc, config.Default(), th)
	e.SetSize(40, 1)

	e.SetMarks([]Mark{{Row: 0, Span: highlight.Span{Start: 1, End: 3, Kind: highlight.Match}}})
	if got, want := e.View(), "  1 B[oo]K"; got != want {
		t.Errorf("View() with a mark = %q, want %q", got, want)
	}
	e.SetMarks(nil)
	if got, want := e.View(), "  1 BOOK"; got != want {
		t.Errorf("View() without marks = %q, want %q", got, want)
	}
}
