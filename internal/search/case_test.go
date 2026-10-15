package search

import (
	"slices"
	"testing"
)

func TestSmartCase(t *testing.T) {
	text := lines{"book", "Book", "BOOK", "ÄRGER"}
	tests := []struct {
		query string
		want  []int
	}{
		{"book", []int{0, 1, 2}},
		{"Book", []int{1}},
		{"BOOK", []int{2}},
		{"ärger", []int{3}},
		{"Ärger", nil},
	}
	for _, m := range matchers {
		for _, tt := range tests {
			if got := rows(m.Find(text, tt.query)); !slices.Equal(got, tt.want) {
				t.Errorf("%s: Find(%q) found rows %v, want %v", m.Name(), tt.query, got, tt.want)
			}
		}
	}
}
