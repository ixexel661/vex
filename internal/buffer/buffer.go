package buffer

import (
	"strings"
	"unicode/utf8"
)

// Position is a place in the text. Col counts runes, not bytes.
type Position struct {
	Row, Col int
}

// Before reports whether p comes before q in the text.
func (p Position) Before(q Position) bool {
	return p.Row < q.Row || p.Row == q.Row && p.Col < q.Col
}

type Buffer struct {
	lines   []string
	version int
}

func New(text string) *Buffer {
	return &Buffer{lines: strings.Split(text, "\n")}
}

func (b *Buffer) Text() string {
	return strings.Join(b.lines, "\n")
}

func (b *Buffer) SetText(text string) {
	b.lines = strings.Split(text, "\n")
	b.version++
}

func (b *Buffer) Line(row int) string {
	return b.lines[row]
}

func (b *Buffer) LineCount() int {
	return len(b.lines)
}

func (b *Buffer) LineLen(row int) int {
	return utf8.RuneCountInString(b.lines[row])
}

// Version changes with every edit, so callers can cheaply detect changes.
func (b *Buffer) Version() int {
	return b.version
}

func (b *Buffer) Clamp(p Position) Position {
	p.Row = min(max(p.Row, 0), len(b.lines)-1)
	p.Col = min(max(p.Col, 0), b.LineLen(p.Row))
	return p
}

func (b *Buffer) Prev(p Position) Position {
	switch {
	case p.Col > 0:
		return Position{Row: p.Row, Col: p.Col - 1}
	case p.Row > 0:
		return Position{Row: p.Row - 1, Col: b.LineLen(p.Row - 1)}
	default:
		return p
	}
}

func (b *Buffer) Next(p Position) Position {
	switch {
	case p.Col < b.LineLen(p.Row):
		return Position{Row: p.Row, Col: p.Col + 1}
	case p.Row < len(b.lines)-1:
		return Position{Row: p.Row + 1, Col: 0}
	default:
		return p
	}
}
