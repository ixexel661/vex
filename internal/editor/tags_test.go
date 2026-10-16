package editor

import (
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/theme"
)

// newTagEditor returns an editor with linked editing on, as in insert mode.
// The cursor is at the "|" in text.
func newTagEditor(t *testing.T, fileType filetype.Type, text string) *Editor {
	t.Helper()
	e := newEditorAt(t, fileType, text)
	e.SetLinkedEditing(true)
	return e
}

func inserting(text string) func(*Editor) {
	return func(e *Editor) { e.InsertText(text) }
}

func TestLinkedEditing(t *testing.T) {
	tests := []struct {
		name string
		text string
		edit func(*Editor)
		want string
	}{
		{"type in the start name", "<bo|ok>x</book>", inserting("x"), "<box|ok>x</boxok>"},
		{"type in the end name", "<book>x</bo|ok>", inserting("x"), "<boxok>x</box|ok>"},
		{"type over several lines", "<bo|ok>\n  x\n</book>", inserting("x"), "<box|ok>\n  x\n</boxok>"},
		{"type in the end name over several lines", "<book>\n  x\n</bo|ok>", inserting("x"), "<boxok>\n  x\n</box|ok>"},
		{"type at the end of the name", "<book|>x</book>", inserting("s"), "<books|>x</books>"},
		{"paste a name", "<|book>x</book>", inserting("my-"), "<my-|book>x</my-book>"},
		{"backspace", "<bo|ok>x</book>", (*Editor).DeleteBackward, "<b|ok>x</bok>"},
		{"backspace in the end name", "<book>x</bo|ok>", (*Editor).DeleteBackward, "<bok>x</b|ok>"},
		{"delete", "<bo|ok>x</book>", (*Editor).DeleteForward, "<bo|k>x</bok>"},
		{"delete the whole name", "<|a>x</a>", (*Editor).DeleteForward, "<|>x</>"},
		{"type into an empty name", "<|>x</>", inserting("b"), "<b|>x</b>"},
		{"backspace after <", "<|book>x</book>", (*Editor).DeleteBackward, "|book>x</book>"},
		{"delete after the name", "<book|>x</book>", (*Editor).DeleteForward, "<book|x</book>"},
		{"space at the end of the name", "<book|>x</book>", inserting(" "), "<book |>x</book>"},
		{"newline", "<bo|ok>x</book>", (*Editor).Newline, "<bo\n|ok>x</book>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTagEditor(t, xml.Type(), tt.text)
			tt.edit(e)
			if got := withCursor(e); got != tt.want {
				t.Errorf("text = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLinkedEditingOff(t *testing.T) {
	e := newTagEditor(t, xml.Type(), "<bo|ok>x</book>")
	e.SetLinkedEditing(false)
	e.InsertText("x")
	if got, want := withCursor(e), "<box|ok>x</book>"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
}

func TestViewMarksTagPair(t *testing.T) {
	brackets := func(open, close string) lipgloss.Style {
		return lipgloss.NewStyle().Transform(func(s string) string { return open + s + close })
	}
	th := theme.Theme{Syntax: map[highlight.Kind]lipgloss.Style{
		highlight.LinkedTag: brackets("[", "]"),
		highlight.Match:     brackets("{", "}"),
	}}
	e := newTagEditor(t, xml.Type(), "<bo|ok>x</book>")
	e.Configure(config.Default(), th)
	if got, want := e.View(), "  1 <[book]>x</[book]>"; got != want {
		t.Errorf("View() = %q, want %q", got, want)
	}
	e.SetMarks([]Mark{{Row: 0, Span: highlight.Span{Start: 1, End: 3, Kind: highlight.Match}}})
	if got, want := e.View(), "  1 <{bo}[ok]>x</[book]>"; got != want {
		t.Errorf("View() with a search match = %q, want %q", got, want)
	}
}

func TestTagPairFollowsEdits(t *testing.T) {
	e := newTagEditor(t, xml.Type(), "\n<|a>x</a>")
	if _, _, ok := e.pair(); !ok {
		t.Fatal("no pair before the edit")
	}
	e.MoveTo(buffer.Position{})
	e.InsertText("<!--")
	e.MoveTo(buffer.Position{Row: 1, Col: 1})
	if _, _, ok := e.pair(); ok {
		t.Error("the tags still pair after they were commented out")
	}
}

func TestCloseTag(t *testing.T) {
	e := newTagEditor(t, xml.Type(), "<p>\n  <title|\n</p>")
	e.CloseTag()
	if got, want := withCursor(e), "<p>\n  <title>|</title>\n</p>"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
}

func TestCloseTagWithoutTags(t *testing.T) {
	e := newTagEditor(t, filetype.Plain(), "<a|")
	e.CloseTag()
	if got, want := withCursor(e), "<a>|"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
}

func TestDeletingNothingKeepsDocumentUnchanged(t *testing.T) {
	e := newEditor(t, "ab", 40, 10)
	e.DeleteBackward()
	e.MoveBottom()
	e.DeleteForward()
	if e.Document().Modified() {
		t.Error("deleting at the ends of the document modified it")
	}
}
