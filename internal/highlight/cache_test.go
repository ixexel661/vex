package highlight

import (
	"strings"
	"testing"
)

type lines []string

func (l lines) Line(row int) string { return l[row] }

// toggler marks lines as Comment between lines containing "*".
type toggler struct{ calls int }

func (t *toggler) Highlight(line []rune, state State) ([]Span, State) {
	t.calls++
	if strings.ContainsRune(string(line), '*') {
		state = 1 - state
	}
	if state == 1 {
		return []Span{{Start: 0, End: len(line), Kind: Comment}}, state
	}
	return nil, state
}

func TestCacheCarriesStateAcrossLines(t *testing.T) {
	text := lines{"a", "*", "b", "*", "c"}
	c := NewCache(&toggler{})
	for row, want := range []bool{false, true, true, false, false} {
		got := len(c.Spans(text, row)) > 0
		if got != want {
			t.Errorf("row %d highlighted = %v, want %v", row, got, want)
		}
	}
}

func TestCacheInvalidate(t *testing.T) {
	text := lines{"a", "*", "b", "c"}
	h := &toggler{}
	c := NewCache(h)
	c.Spans(text, 3)

	text[1] = "x"
	c.Invalidate(1)
	if spans := c.Spans(text, 3); len(spans) != 0 {
		t.Errorf("row 3 still highlighted after the comment start was removed")
	}

	h.calls = 0
	c.Spans(text, 3)
	if h.calls != 1 {
		t.Errorf("cached lookup highlighted %d lines, want 1", h.calls)
	}
}

func TestCacheState(t *testing.T) {
	text := lines{"a", "*", "b", "*"}
	c := NewCache(&toggler{})
	// Out of order, so some states come from the cache.
	for _, tt := range []struct {
		row  int
		want State
	}{{2, 1}, {0, 0}, {3, 1}, {1, 0}} {
		if got := c.State(text, tt.row); got != tt.want {
			t.Errorf("State(%d) = %d, want %d", tt.row, got, tt.want)
		}
	}
}

func TestCacheWithoutHighlighter(t *testing.T) {
	c := NewCache(nil)
	if spans := c.Spans(lines{"a"}, 0); spans != nil {
		t.Errorf("Spans() = %v, want nil", spans)
	}
	if state := c.State(lines{"a", "b"}, 1); state != 0 {
		t.Errorf("State() = %d, want 0", state)
	}
}
