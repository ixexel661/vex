package editor

import "github.com/ixexel661/vex/internal/highlight"

// Mark highlights a part of a row on top of the syntax highlighting, e.g. a
// search match.
type Mark struct {
	Row  int
	Span highlight.Span
}

// SetMarks replaces the marks shown in the text.
func (e *Editor) SetMarks(marks []Mark) {
	e.marks = map[int][]highlight.Span{}
	for _, m := range marks {
		e.marks[m.Row] = append(e.marks[m.Row], m.Span)
	}
}
