package editor

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/theme"
)

// problem marks the runes from "from" to "to" of row.
func problem(row, from, to int, message string) filetype.Problem {
	return filetype.Problem{
		Range:   buffer.Range{Start: buffer.Position{Row: row, Col: from}, End: buffer.Position{Row: row, Col: to}},
		Message: message,
	}
}

// problemsIn shows the text with each problem in brackets.
func problemsIn(e *Editor) string {
	lines := strings.Split(e.Document().Buffer.Text(), "\n")
	problems := e.Problems()
	for i := len(problems) - 1; i >= 0; i-- {
		r := problems[i].Range
		line := []rune(lines[r.Start.Row])
		lines[r.Start.Row] = string(line[:r.Start.Col]) + "[" + string(line[r.Start.Col:r.End.Col]) + "]" + string(line[r.End.Col:])
	}
	return strings.Join(lines, "\n")
}

func TestViewMarksProblems(t *testing.T) {
	transform := func(open, close string) lipgloss.Style {
		return lipgloss.NewStyle().Transform(func(s string) string { return open + s + close })
	}
	th := theme.Theme{
		Syntax: map[highlight.Kind]lipgloss.Style{highlight.Problem: transform("[", "]")},
		Error:  transform("!", ""),
	}
	e := newEditor(t, "<a>\n  <b/>\n</a>", 40, 3)
	e.Configure(config.Default(), th)
	e.SetProblems([]filetype.Problem{problem(1, 3, 4, "b is wrong")})
	want := "  1 <a>\n!  2   <[b]/>\n  3 </a>"
	if got := e.View(); got != want {
		t.Errorf("View() =\n%s\nwant\n%s", got, want)
	}
}

func TestProblemsMoveWithTheText(t *testing.T) {
	e := newEditorAt(t, filetype.Plain(), "|<a>\n<b/><c/>\n</a>")
	e.SetProblems([]filetype.Problem{problem(1, 1, 2, "b"), problem(1, 5, 6, "c"), problem(2, 2, 3, "a")})
	steps := []struct {
		name string
		edit func(e *Editor)
		want string
	}{
		{"type before them", func(e *Editor) {
			e.MoveTo(buffer.Position{Row: 1})
			e.InsertText("xx")
		}, "<a>\nxx<[b]/><[c]/>\n</[a]>"},
		{"new line above", func(e *Editor) {
			e.MoveTo(buffer.Position{Row: 0, Col: 3})
			e.Newline()
		}, "<a>\n\nxx<[b]/><[c]/>\n</[a]>"},
		{"line break between them", func(e *Editor) {
			e.MoveTo(buffer.Position{Row: 2, Col: 6})
			e.Newline()
		}, "<a>\n\nxx<[b]/>\n<[c]/>\n</[a]>"},
		{"delete a marked name", func(e *Editor) {
			e.MoveTo(buffer.Position{Row: 2, Col: 4})
			e.DeleteBackward()
		}, "<a>\n\nxx<[]/>\n<[c]/>\n</[a]>"},
		{"join lines", func(e *Editor) {
			e.MoveTo(buffer.Position{Row: 1})
			e.DeleteBackward()
		}, "<a>\nxx<[]/>\n<[c]/>\n</[a]>"},
	}
	for _, step := range steps {
		step.edit(e)
		if got := problemsIn(e); got != step.want {
			t.Fatalf("%s: problems in\n%s\nwant\n%s", step.name, got, step.want)
		}
	}
}

func TestProblemAt(t *testing.T) {
	e := newEditor(t, "a\nb\nc\nd", 40, 4)
	e.SetProblems([]filetype.Problem{problem(1, 0, 1, "first"), problem(1, 0, 1, "second"), problem(3, 0, 1, "last")})
	for row, want := range []string{"", "first", "", "last"} {
		p, ok := e.ProblemAt(row)
		if p.Message != want || ok != (want != "") {
			t.Errorf("ProblemAt(%d) = %q, %v, want %q", row, p.Message, ok, want)
		}
	}
	e.SetText("x")
	if len(e.Problems()) != 0 {
		t.Error("the problems of the old text stay after SetText")
	}
}
