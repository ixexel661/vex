package search

import (
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
)

func TestSessionSelectsFirstMatchFromOrigin(t *testing.T) {
	s := NewSession(lines{"x", "ax", "x"}, []Matcher{Text{}})
	s.SetQuery("x")
	tests := []struct {
		name   string
		origin buffer.Position
		want   int
	}{
		{"match at the origin", buffer.Position{Row: 0, Col: 0}, 0},
		{"match after the origin", buffer.Position{Row: 0, Col: 1}, 1},
		{"wraps around", buffer.Position{Row: 2, Col: 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.Start(tt.origin)
			if got := s.Selected(); got != tt.want {
				t.Errorf("Selected() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSessionMoveWrapsAround(t *testing.T) {
	s := NewSession(lines{"x x x"}, []Matcher{Text{}})
	s.SetQuery("x")
	for _, step := range []struct{ delta, want int }{{1, 1}, {1, 2}, {1, 0}, {-1, 2}, {-1, 1}} {
		s.Move(step.delta)
		if got := s.Selected(); got != step.want {
			t.Fatalf("Move(%d) selected %d, want %d", step.delta, got, step.want)
		}
	}
}

func TestSessionNextMatcherKeepsQuery(t *testing.T) {
	s := NewSession(lines{"<book>", "notebook"}, []Matcher{Text{}, Fuzzy{}})
	s.Start(buffer.Position{Row: 1})
	s.SetQuery("book")
	if got := s.Matches()[s.Selected()].Row; got != 1 {
		t.Fatalf("text search selected row %d, want the first match from the origin in row 1", got)
	}

	s.NextMatcher()
	if s.Matcher().Name() != "Fuzzy" || s.Query() != "book" {
		t.Fatalf("matcher %s with query %q, want Fuzzy with %q", s.Matcher().Name(), s.Query(), "book")
	}
	if got := s.Matches()[s.Selected()].Row; got != 0 {
		t.Errorf("fuzzy search selected row %d, want the best match in row 0", got)
	}

	s.NextMatcher()
	if s.Matcher().Name() != "Text" {
		t.Errorf("matcher %s after the last one, want Text", s.Matcher().Name())
	}
}

func TestSessionWithoutMatches(t *testing.T) {
	s := NewSession(lines{"abc"}, []Matcher{Text{}})
	for _, query := range []string{"", "x"} {
		s.SetQuery(query)
		if s.Selected() != -1 || len(s.Matches()) != 0 {
			t.Errorf("query %q: selected %d of %v, want -1 and no matches", query, s.Selected(), s.Matches())
		}
		s.Move(1)
		if s.Selected() != -1 {
			t.Errorf("query %q: Move() selected %d, want -1", query, s.Selected())
		}
	}
}
