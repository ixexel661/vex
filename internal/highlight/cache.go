package highlight

type Lines interface {
	Line(row int) string
}

// Cache remembers the state at the start of each line, so after an edit only
// the lines from the edited one onwards have to be highlighted again.
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

func (c *Cache) Spans(lines Lines, row int) []Span {
	if c.highlighter == nil {
		return nil
	}
	for len(c.states) <= row {
		last := len(c.states) - 1
		_, next := c.highlighter.Highlight([]rune(lines.Line(last)), c.states[last])
		c.states = append(c.states, next)
	}
	spans, _ := c.highlighter.Highlight([]rune(lines.Line(row)), c.states[row])
	return spans
}
