package search

import (
	"reflect"
	"testing"
)

// at returns a match of the runes from start up to end of row.
func at(row, start, end int) Match {
	return Match{Row: row, Ranges: []Range{{Start: start, End: end}}}
}

func TestTextFindsEveryOccurrence(t *testing.T) {
	tests := []struct {
		name  string
		lines lines
		query string
		want  []Match
	}{
		{"several per line", lines{"<a><a>"}, "<a>", []Match{at(0, 0, 3), at(0, 3, 6)}},
		{"no overlaps", lines{"aaaa"}, "aa", []Match{at(0, 0, 2), at(0, 2, 4)}},
		{"rune positions", lines{"äöü äöü"}, "öü", []Match{at(0, 1, 3), at(0, 5, 7)}},
		{"in text order", lines{"x", "y", "yx"}, "x", []Match{at(0, 0, 1), at(2, 1, 2)}},
		{"no match", lines{"abc"}, "abd", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Text{}).Find(tt.lines, tt.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
