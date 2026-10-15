package search

import "github.com/ixexel661/vex/internal/buffer"

// Lines is the text to search, e.g. a buffer.
type Lines interface {
	Line(row int) string
	LineCount() int
}

// Range marks the runes from Start up to, but not including, End of a line.
type Range struct {
	Start, End int
}

// Match is a hit in one line. Ranges are the parts of the line that match
// the query, from left to right.
type Match struct {
	Row    int
	Ranges []Range
}

// Position returns where the match starts.
func (m Match) Position() buffer.Position {
	p := buffer.Position{Row: m.Row}
	if len(m.Ranges) > 0 {
		p.Col = m.Ranges[0].Start
	}
	return p
}

// Matcher is a way to search, e.g. for the exact text. An empty query
// matches nothing.
type Matcher interface {
	Name() string
	// Ranked reports whether the matches are sorted by relevance instead of
	// by position.
	Ranked() bool
	Find(lines Lines, query string) []Match
}
