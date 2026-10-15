package search

import (
	"reflect"
	"slices"
	"testing"
)

func TestFuzzyRanges(t *testing.T) {
	tests := []struct {
		name, line, query string
		want              []Range
	}{
		{"adjacent runes join", "<title>", "ttl", []Range{{1, 2}, {3, 5}}},
		{"the latest start wins", "x-xyz", "xyz", []Range{{2, 5}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fuzzy{}.Find(lines{tt.line}, tt.query)
			if len(got) != 1 || !reflect.DeepEqual(got[0].Ranges, tt.want) {
				t.Errorf("Find(%q) in %q = %v, want ranges %v", tt.query, tt.line, got, tt.want)
			}
		})
	}
}

func TestFuzzyRanksWordStartsAndRuns(t *testing.T) {
	text := lines{
		"a big old oak key", // scattered over word starts
		"<book>",            // a word start and a run
		"nothing",           // no match
		"notebook",          // a run
	}
	got := rows(Fuzzy{}.Find(text, "book"))
	if want := []int{1, 3, 0}; !slices.Equal(got, want) {
		t.Errorf("found rows %v, want %v", got, want)
	}
}

func TestFuzzyKeepsOrderOfEqualMatches(t *testing.T) {
	// Enough lines that an unstable sort would mix them up.
	var text lines
	var best, rest []int
	for row := range 40 {
		if row%3 == 0 {
			text = append(text, "<book>")
			best = append(best, row)
		} else {
			text = append(text, "notebook")
			rest = append(rest, row)
		}
	}
	got := rows(Fuzzy{}.Find(text, "book"))
	if want := slices.Concat(best, rest); !slices.Equal(got, want) {
		t.Errorf("found rows %v, want %v", got, want)
	}
}
