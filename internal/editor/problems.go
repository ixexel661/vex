package editor

import (
	"slices"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/highlight"
)

// SetProblems marks problems in the text, e.g. from a schema check. They
// have to be sorted by position. The editor keeps them where their text is
// as the text changes, until the next call.
func (e *Editor) SetProblems(problems []filetype.Problem) {
	e.problems = problems
}

// Problems returns the problems in the text, sorted by position.
func (e *Editor) Problems() []filetype.Problem {
	return e.problems
}

// ProblemAt returns the first problem that starts in row.
func (e *Editor) ProblemAt(row int) (filetype.Problem, bool) {
	i := e.firstProblemIn(row)
	if i < len(e.problems) && e.problems[i].Range.Start.Row == row {
		return e.problems[i], true
	}
	return filetype.Problem{}, false
}

// firstProblemIn returns the index of the first problem in row or after it.
func (e *Editor) firstProblemIn(row int) int {
	i, _ := slices.BinarySearchFunc(e.problems, row, func(p filetype.Problem, row int) int {
		return p.Range.Start.Row - row
	})
	return i
}

// problemSpans marks the problems that start in row.
func (e *Editor) problemSpans(row int) []highlight.Span {
	var spans []highlight.Span
	for _, p := range e.problems[e.firstProblemIn(row):] {
		if p.Range.Start.Row != row {
			break
		}
		end := p.Range.End.Col
		if p.Range.End.Row > row {
			end = len([]rune(e.buffer().Line(row)))
		}
		spans = append(spans, highlight.Span{Start: p.Range.Start.Col, End: end, Kind: highlight.Problem})
	}
	return spans
}

// moveProblems keeps the problems at their text after the text from start
// to end was replaced by text that ends at newEnd.
func (e *Editor) moveProblems(start, end, newEnd buffer.Position) {
	if start == end && start == newEnd {
		return
	}
	for i := range e.problems {
		r := &e.problems[i].Range
		r.Start = moved(r.Start, start, end, newEnd)
		r.End = moved(r.End, start, end, newEnd)
	}
}

// moved returns where p is after the text from start to end was replaced
// by text that ends at newEnd. A position in the replaced text goes to its
// start.
func moved(p, start, end, newEnd buffer.Position) buffer.Position {
	switch {
	case p.Before(start):
		return p
	case p.Before(end):
		return start
	case p.Row == end.Row:
		return buffer.Position{Row: newEnd.Row, Col: newEnd.Col + p.Col - end.Col}
	default:
		return buffer.Position{Row: p.Row + newEnd.Row - end.Row, Col: p.Col}
	}
}
