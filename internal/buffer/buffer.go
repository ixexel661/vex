package buffer

import (
	"cmp"
	"strings"
	"sync/atomic"
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

// Compare returns -1 if p comes before q, 1 if it comes after q and 0 if
// they are the same.
func (p Position) Compare(q Position) int {
	return cmp.Or(cmp.Compare(p.Row, q.Row), cmp.Compare(p.Col, q.Col))
}

// Range is the text from Start up to, but not including, End.
type Range struct {
	Start, End Position
}

type Buffer struct {
	lines   []string
	version int
	readers *atomic.Int32 // snapshots that still read lines
}

func New(text string) *Buffer {
	return &Buffer{lines: split(text, false)}
}

// NewCRLF is New for text with "\r\n" line endings, whose "\r" goes with the
// line break.
func NewCRLF(text string) *Buffer {
	return &Buffer{lines: split(text, true)}
}

func (b *Buffer) Text() string {
	return strings.Join(b.lines, "\n")
}

func (b *Buffer) SetText(text string) {
	b.lines = split(text, false)
	b.readers = nil
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
