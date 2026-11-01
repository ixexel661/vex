package highlight

import "slices"

type Lines interface {
	Line(row int) string
	LineCount() int
}

// resyncLimit is how many rows Edit highlights again at most. If the states
// still differ after them, the rows below are highlighted when needed.
const resyncLimit = 1000

// Cache remembers the state at the start of each line, so after an edit only
// the lines whose state changes have to be highlighted again.
type Cache struct {
	highlighter Highlighter
	states      []State
}

func NewCache(h Highlighter) *Cache {
	return &Cache{highlighter: h, states: []State{0}}
}

// Invalidate forgets the states after row. The state at the start of row
// only depends on the lines above it, so it stays valid.
func (c *Cache) Invalidate(row int) {
	if row+1 < len(c.states) {
		c.states = c.states[:row+1]
	}
}

// Edit updates the states after the rows from row to row+removed were
// replaced by the rows from row to row+added. The states of the rows below
// move with them and stay, as soon as the edited rows end in the state
// they ended in before.
func (c *Cache) Edit(lines Lines, row, removed, added int) {
	if c.highlighter == nil {
		return
	}
	tail := row + removed + 1
	if tail >= len(c.states) {
		c.Invalidate(row)
		return
	}
	c.states = slices.Replace(c.states, row+1, tail, make([]State, added)...)
	for r := row; r < row+resyncLimit; r++ {
		if r+1 >= len(c.states) {
			return
		}
		_, next := c.highlighter.Highlight([]rune(lines.Line(r)), c.states[r])
		if r >= row+added && c.states[r+1] == next {
			return
		}
		c.states[r+1] = next
	}
	c.Invalidate(row + resyncLimit)
}

func (c *Cache) Spans(lines Lines, row int) []Span {
	if c.highlighter == nil {
		return nil
	}
	spans, _ := c.highlighter.Highlight([]rune(lines.Line(row)), c.State(lines, row))
	return spans
}

// State returns the state at the start of row.
func (c *Cache) State(lines Lines, row int) State {
	if c.highlighter == nil {
		return 0
	}
	for len(c.states) <= row {
		last := len(c.states) - 1
		_, next := c.highlighter.Highlight([]rune(lines.Line(last)), c.states[last])
		c.states = append(c.states, next)
	}
	return c.states[row]
}

// Clone returns a copy of the cache, e.g. for another goroutine.
func (c *Cache) Clone() *Cache {
	return &Cache{highlighter: c.highlighter, states: slices.Clone(c.states)}
}

// Known returns the last row whose state the cache knows.
func (c *Cache) Known() int {
	return len(c.states) - 1
}

// Extend adds the states that States computed from row on, if the cache
// still reaches row and does not know them yet.
func (c *Cache) Extend(row int, states []State) {
	known := len(c.states) - row - 1
	if known < 0 || known >= len(states) {
		return
	}
	c.states = append(c.states, states[known:]...)
}

// States highlights the lines from row on, starting in state, and returns
// the states at the start of the rows after row. It only reads lines, so it
// can run in another goroutine on a snapshot of them.
func States(h Highlighter, lines Lines, row int, state State) []State {
	states := make([]State, 0, max(lines.LineCount()-row-1, 0))
	for r := row; r < lines.LineCount()-1; r++ {
		_, state = h.Highlight([]rune(lines.Line(r)), state)
		states = append(states, state)
	}
	return states
}
