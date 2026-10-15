package search

import "slices"

// Text finds every occurrence of the query, in the order they appear.
type Text struct{}

func (Text) Name() string {
	return "Text"
}

func (Text) Ranked() bool {
	return false
}

func (Text) Find(lines Lines, query string) []Match {
	pattern, ignoreCase := prepare(query)
	if len(pattern) == 0 {
		return nil
	}
	var matches []Match
	for row := range lines.LineCount() {
		for _, col := range occurrences(normalize(lines.Line(row), ignoreCase), pattern) {
			matches = append(matches, Match{Row: row, Ranges: []Range{{Start: col, End: col + len(pattern)}}})
		}
	}
	return matches
}

// occurrences returns where pattern starts in line. The search goes on after
// the end of each occurrence, so they do not overlap.
func occurrences(line, pattern []rune) []int {
	var starts []int
	for col := 0; col+len(pattern) <= len(line); {
		if slices.Equal(line[col:col+len(pattern)], pattern) {
			starts = append(starts, col)
			col += len(pattern)
		} else {
			col++
		}
	}
	return starts
}
