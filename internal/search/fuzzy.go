package search

import (
	"cmp"
	"slices"
	"unicode"
)

const (
	matchScore       = 16
	wordStartBonus   = 8
	consecutiveBonus = 8
	gapPenalty       = 1
)

// Fuzzy finds the lines that contain the runes of the query in order, with
// anything in between, and ranks them by how well they match.
type Fuzzy struct{}

func (Fuzzy) Name() string {
	return "Fuzzy"
}

func (Fuzzy) Ranked() bool {
	return true
}

func (Fuzzy) Find(lines Lines, query string) []Match {
	pattern, ignoreCase := prepare(query)
	if len(pattern) == 0 {
		return nil
	}
	var found []scored
	for row := range lines.LineCount() {
		line := normalize(lines.Line(row), ignoreCase)
		if positions := locate(line, pattern); positions != nil {
			found = append(found, scored{Match{Row: row, Ranges: ranges(positions)}, score(line, positions)})
		}
	}
	// The sort is stable, so of two equally good lines the earlier one wins.
	slices.SortStableFunc(found, func(a, b scored) int {
		return cmp.Compare(b.score, a.score)
	})
	var matches []Match
	for _, f := range found {
		matches = append(matches, f.match)
	}
	return matches
}

type scored struct {
	match Match
	score int
}

// locate returns the positions of the runes of pattern in line, or nil if
// line does not contain them in order. Like fzf v1, it looks forward for the
// end of the first match, backward from there for the latest start, and then
// takes each rune as early as possible from that start. This keeps the
// match short.
func locate(line, pattern []rune) []int {
	end := forward(line, pattern)
	if end < 0 {
		return nil
	}
	start := backward(line[:end+1], pattern)
	positions := make([]int, 0, len(pattern))
	for p := start; len(positions) < len(pattern); p++ {
		if line[p] == pattern[len(positions)] {
			positions = append(positions, p)
		}
	}
	return positions
}

// forward returns the index of the rune that completes the first match of
// pattern in line, or -1 if there is none.
func forward(line, pattern []rune) int {
	i := 0
	for p, r := range line {
		if r == pattern[i] {
			i++
			if i == len(pattern) {
				return p
			}
		}
	}
	return -1
}

// backward returns the latest index a match of pattern in line starts at,
// or -1 if there is none.
func backward(line, pattern []rune) int {
	i := len(pattern) - 1
	for p := len(line) - 1; p >= 0; p-- {
		if line[p] == pattern[i] {
			if i == 0 {
				return p
			}
			i--
		}
	}
	return -1
}

// score rates a match. Each matched rune counts, more so at the start of a
// word or right after the previous one, and each rune skipped in between
// costs a little.
func score(line []rune, positions []int) int {
	total := 0
	for i, p := range positions {
		total += matchScore
		if p == 0 || !isWordRune(line[p-1]) {
			total += wordStartBonus
		}
		if i == 0 {
			continue
		}
		if gap := p - positions[i-1] - 1; gap == 0 {
			total += consecutiveBonus
		} else {
			total -= gap * gapPenalty
		}
	}
	return total
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ranges joins adjacent positions.
func ranges(positions []int) []Range {
	var out []Range
	for _, p := range positions {
		if n := len(out); n > 0 && out[n-1].End == p {
			out[n-1].End++
		} else {
			out = append(out, Range{Start: p, End: p + 1})
		}
	}
	return out
}
