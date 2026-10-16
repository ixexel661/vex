package xml

import (
	"iter"
	"slices"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
)

// Tags reads tags with the lexer of the highlighter, so tags in comments,
// CDATA and attribute values don't count. It only reads the lines it needs.
type Tags struct{}

type partKind int

const (
	startName   partKind = iota + 1 // the name after "<"
	endName                         // the name after "</"
	tagEnd                          // ">"
	emptyTagEnd                     // "/>"
)

// part is a tag name, which may be empty, or the end of a tag: the runes
// from start up to end of a row.
type part struct {
	kind       partKind
	row        int
	start, end int
	name       string
}

func (p part) isName() bool {
	return p.kind == startName || p.kind == endName
}

func (p part) textRange() buffer.Range {
	return buffer.Range{
		Start: buffer.Position{Row: p.row, Col: p.start},
		End:   buffer.Position{Row: p.row, Col: p.end},
	}
}

// tag is a tag name and the end that follows it. A tag that is not finished,
// e.g. while it is typed, has no end.
type tag struct {
	name, end part
}

func (Tags) Pair(src filetype.Source, p buffer.Position) (name, other buffer.Range, ok bool) {
	at, ok := nameAt(src, p)
	if !ok {
		return name, other, false
	}
	tags := forward(src, p.Row)
	if at.kind == endName {
		tags = backward(src, p.Row)
	}
	to, ok := partner(tags, at)
	return at.textRange(), to.textRange(), ok
}

func (Tags) IsName(text string) bool {
	for _, r := range text {
		if !isNameRune(r) {
			return false
		}
	}
	return true
}

func (Tags) Close(src filetype.Source, p buffer.Position) (string, bool) {
	for t := range backward(src, p.Row) {
		if t.name.row == p.Row && t.name.start >= p.Col {
			continue
		}
		endsAtP := t.end.kind == tagEnd && t.end.row == p.Row && t.end.end == p.Col
		if t.name.kind != startName || t.name.name == "" || !endsAtP {
			return "", false
		}
		if _, ok := partner(forward(src, t.name.row), t.name); ok {
			return "", false
		}
		return "</" + t.name.name + ">", true
	}
	return "", false
}

// nameAt returns the tag name that p is in or right before or after.
func nameAt(src filetype.Source, p buffer.Position) (part, bool) {
	for _, pt := range lineParts(src, p.Row) {
		if pt.isName() && pt.start <= p.Col && p.Col <= pt.end {
			return pt, true
		}
	}
	return part{}, false
}

// partner finds the other tag of the element that at names. Tags of the kind
// of at go a level deeper, the others go back up. The tags named like at
// have a level of their own, so a child without an end tag doesn't break the
// pair. Going above the level of at leaves the surrounding element, where
// the search stops.
func partner(tags iter.Seq[tag], at part) (part, bool) {
	started := false
	same, level := 0, 0
	for t := range tags {
		switch {
		case !started:
			if t.name == at {
				if t.end.kind != tagEnd {
					return part{}, false
				}
				started = true
			}
		case t.end.kind != tagEnd:
		case t.name.kind == at.kind:
			level++
			if t.name.name == at.name {
				same++
			}
		case t.name.name == at.name && same == 0:
			return t.name, true
		default:
			level--
			if t.name.name == at.name {
				same--
			}
			if level < 0 {
				return part{}, false
			}
		}
	}
	return part{}, false
}

// forward returns the tags from the start of row to the end of the document.
func forward(src filetype.Source, row int) iter.Seq[tag] {
	return func(yield func(tag) bool) {
		var name part
		for p := range partsFrom(src, row, 1) {
			switch {
			case p.isName():
				if name.kind != 0 && !yield(tag{name: name}) {
					return
				}
				name = p
			case name.kind != 0:
				if !yield(tag{name: name, end: p}) {
					return
				}
				name = part{}
			}
		}
		if name.kind != 0 {
			yield(tag{name: name})
		}
	}
}

// backward returns the tags from the end of row to the start of the
// document, the last one first.
func backward(src filetype.Source, row int) iter.Seq[tag] {
	return func(yield func(tag) bool) {
		var end part
		for p := range partsFrom(src, row, -1) {
			if !p.isName() {
				end = p
				continue
			}
			if !yield(tag{name: p, end: end}) {
				return
			}
			end = part{}
		}
	}
}

// partsFrom returns the parts of the rows from row on, going down for a step
// of 1 and up for -1. Going up, the parts of each row come last to first.
func partsFrom(src filetype.Source, row, step int) iter.Seq[part] {
	return func(yield func(part) bool) {
		for r := row; r >= 0 && r < src.LineCount(); r += step {
			parts := lineParts(src, r)
			if step < 0 {
				slices.Reverse(parts)
			}
			for _, p := range parts {
				if !yield(p) {
					return
				}
			}
		}
	}
}

// lineParts lexes a row from the state the highlighter has at its start.
func lineParts(src filetype.Source, row int) []part {
	parts := lex([]rune(src.Line(row)), src.State(row)).parts
	for i := range parts {
		parts[i].row = row
	}
	return parts
}
