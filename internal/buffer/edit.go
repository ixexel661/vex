package buffer

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// Insert adds text at p and returns the position right after the inserted text.
func (b *Buffer) Insert(p Position, text string) Position {
	before, after := b.split(p)
	parts := strings.Split(text, "\n")
	last := len(parts) - 1
	parts[0] = before + parts[0]
	end := Position{Row: p.Row + last, Col: utf8.RuneCountInString(parts[last])}
	parts[last] += after
	b.lines = slices.Replace(b.lines, p.Row, p.Row+1, parts...)
	b.version++
	return end
}

// Delete removes the text from "from" up to, but not including, "to".
func (b *Buffer) Delete(from, to Position) {
	if from == to {
		return
	}
	head, _ := b.split(from)
	_, tail := b.split(to)
	b.lines = slices.Replace(b.lines, from.Row, to.Row+1, head+tail)
	b.version++
}

func (b *Buffer) split(p Position) (before, after string) {
	line := []rune(b.lines[p.Row])
	return string(line[:p.Col]), string(line[p.Col:])
}
