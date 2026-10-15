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

func TestCacheWithoutHighlighter(t *testing.T) {
	if spans := NewCache(nil).Spans(lines{"a"}, 0); spans != nil {
		t.Errorf("Spans() = %v, want nil", spans)
	}
}
