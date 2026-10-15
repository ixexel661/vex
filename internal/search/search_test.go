package search

import "testing"

type lines []string

func (l lines) Line(row int) string { return l[row] }
func (l lines) LineCount() int      { return len(l) }

var matchers = []Matcher{Text{}, Fuzzy{}}

// rows returns the row of each match.
func rows(matches []Match) []int {
	var out []int
	for _, m := range matches {
		out = append(out, m.Row)
	}
	return out
}

func TestEmptyQueryMatchesNothing(t *testing.T) {
	for _, m := range matchers {
		if got := m.Find(lines{"abc", ""}, ""); got != nil {
			t.Errorf("%s: Find(\"\") = %v, want nil", m.Name(), got)
		}
	}
}
