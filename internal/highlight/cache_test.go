package highlight

import (
	"slices"
	"strings"
	"testing"
)

type lines []string

func (l lines) Line(row int) string { return l[row] }
func (l lines) LineCount() int      { return len(l) }

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

func manyLines(n int) lines {
	text := make(lines, n)
	for i := range text {
		text[i] = "a"
	}
	return text
}

func TestCacheEditKeepsTheStatesBelow(t *testing.T) {
	text := manyLines(3000)
	text[2000] = "*"
	h := &toggler{}
	c := NewCache(h)
	c.State(text, 2999)

	h.calls = 0
	text[10] = "b"
	c.Edit(text, 10, 0, 0)
	if h.calls != 1 || c.Known() != 2999 {
		t.Errorf("edit without a state change highlighted %d lines and knows %d rows, want 1 and all", h.calls, c.Known())
	}
	if c.State(text, 2500) != 1 {
		t.Error("the state after the comment start is lost")
	}
}

func TestCacheEditStopsAfterTheLimit(t *testing.T) {
	text := manyLines(3000)
	h := &toggler{}
	c := NewCache(h)
	c.State(text, 2999)

	h.calls = 0
	text[10] = "*"
	c.Edit(text, 10, 0, 0)
	if h.calls != resyncLimit || c.Known() != 10+resyncLimit {
		t.Errorf("highlighted %d lines and knows %d rows, want %d and %d", h.calls, c.Known(), resyncLimit, 10+resyncLimit)
	}
	if c.State(text, 2999) != 1 {
		t.Error("the rows after the limit kept their old state")
	}
}

func TestCacheEditMovesTheStatesBelow(t *testing.T) {
	text := lines{"a", "*", "b", "*", "c"}
	c := NewCache(&toggler{})
	c.State(text, 4)

	// "b" becomes three lines, then "*" and the line after it go.
	text = lines{"a", "*", "x", "y", "z", "*", "c"}
	c.Edit(text, 2, 0, 2)
	text = lines{"a", "*", "x", "y", "c"}
	c.Edit(text, 4, 2, 0)
	for row, want := range []State{0, 0, 1, 1, 1} {
		if got := c.State(text, row); got != want {
			t.Errorf("State(%d) = %d, want %d", row, got, want)
		}
	}
}

func TestPrefetchedStates(t *testing.T) {
	text := lines{"a", "*", "b", "*", "c"}
	c := NewCache(&toggler{})
	c.State(text, 1)
	states := States(&toggler{}, text, c.Known(), c.State(text, c.Known()))
	if want := []State{1, 1, 0}; !slices.Equal(states, want) {
		t.Fatalf("States() = %v, want %v", states, want)
	}
	c.State(text, 2) // the cache went on meanwhile
	c.Extend(1, states)
	if c.Known() != 4 || c.State(text, 4) != 0 {
		t.Errorf("Extend: knows %d rows, state of the last = %d, want 4 and 0", c.Known(), c.State(text, 4))
	}

	c.Invalidate(0)
	c.Extend(1, states)
	if c.Known() != 0 {
		t.Errorf("Extend added states to a cache that no longer reaches their row")
	}
}

func TestCacheClone(t *testing.T) {
	text := lines{"a", "*", "b"}
	c := NewCache(&toggler{})
	c.State(text, 2)
	clone := c.Clone()
	c.Edit(lines{"a", "a", "b"}, 1, 0, 0)
	if clone.Known() != 2 || clone.State(text, 2) != 1 {
		t.Errorf("after an edit of the original, the clone knows %d rows and state %d of the last, want 2 and 1", clone.Known(), clone.State(text, 2))
	}
}
