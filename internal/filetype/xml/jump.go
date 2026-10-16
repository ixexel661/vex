package xml

import (
	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
)

func (Tags) NextTag(src filetype.Source, p buffer.Position) (buffer.Position, bool) {
	for t := range forward(src, p.Row) {
		if p.Before(t.open()) {
			return t.open(), true
		}
	}
	return p, false
}

func (Tags) PrevTag(src filetype.Source, p buffer.Position) (buffer.Position, bool) {
	for t := range backward(src, p.Row) {
		if t.open().Before(p) {
			return t.open(), true
		}
	}
	return p, false
}

// Parent goes back from p and counts the elements that end before p, until
// it finds a start tag that is not ended. A start tag without an end, like
// one that goes on below the row of p, counts as a start tag.
func (Tags) Parent(src filetype.Source, p buffer.Position) (buffer.Position, bool) {
	depth := 0
	for t := range backward(src, p.Row) {
		open := t.open()
		switch {
		case !open.Before(p):
		case t.name.kind == endName:
			// An end tag around p belongs to the element around p.
			if t.end.kind != 0 && !p.Before(t.after()) {
				depth++
			}
		case t.end.kind == emptyTagEnd:
			if p.Before(t.after()) {
				return open, true
			}
		case depth == 0:
			return open, true
		default:
			depth--
		}
	}
	return p, false
}

// FirstChild looks at the tag after the start tag of the element at p,
// which is either its first child or its end tag. Outside of all elements,
// the first child is the root element.
func (tags Tags) FirstChild(src filetype.Source, p buffer.Position) (buffer.Position, bool) {
	start, ok := tags.element(src, p)
	if !ok {
		return root(src)
	}
	inside := false
	for t := range forward(src, start.Row) {
		switch {
		case !inside:
			if t.open() == start {
				if t.end.kind != tagEnd {
					return p, false
				}
				inside = true
			}
		case t.name.kind == startName:
			return t.open(), true
		default:
			return p, false
		}
	}
	return p, false
}

// element returns where the start tag of the element at p starts: the one
// that starts at p, or else the one around p.
func (tags Tags) element(src filetype.Source, p buffer.Position) (buffer.Position, bool) {
	for t := range forward(src, p.Row) {
		if p.Before(t.open()) {
			break
		}
		if t.open() == p && t.name.kind == startName {
			return p, true
		}
	}
	return tags.Parent(src, p)
}

func root(src filetype.Source) (buffer.Position, bool) {
	for t := range forward(src, 0) {
		if t.name.kind == startName {
			return t.open(), true
		}
	}
	return buffer.Position{}, false
}

// open returns where the tag starts, at its "<".
func (t tag) open() buffer.Position {
	col := t.name.start - 1
	if t.name.kind == endName {
		col--
	}
	return buffer.Position{Row: t.name.row, Col: col}
}

// after returns where the tag ends, right after its ">" or "/>".
func (t tag) after() buffer.Position {
	return buffer.Position{Row: t.end.row, Col: t.end.end}
}
