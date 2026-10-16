package editor

import (
	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/highlight"
)

// SetLinkedEditing turns linked editing on or off: the names of the start
// and end tag at the cursor are marked and edited together.
func (e *Editor) SetLinkedEditing(on bool) {
	e.linkedEditing = on
}

// CloseTag types ">" and, if that ends a start tag, adds the end tag after
// the cursor.
func (e *Editor) CloseTag() {
	e.InsertText(">")
	tags := e.doc.Type.Tags
	if tags == nil {
		return
	}
	if end, ok := tags.Close(e.source(), e.cursor); ok {
		e.change(buffer.Range{Start: e.cursor, End: e.cursor}, end)
	}
}

// pair returns the tag name at the cursor and the other name of its element.
func (e *Editor) pair() (name, other buffer.Range, ok bool) {
	tags := e.doc.Type.Tags
	if !e.linkedEditing || tags == nil {
		return name, other, false
	}
	return tags.Pair(e.source(), e.cursor)
}

func (e *Editor) pairMarks() map[int][]highlight.Span {
	name, other, ok := e.pair()
	if !ok {
		return nil
	}
	return byRow([]Mark{linkedTag(name), linkedTag(other)})
}

func linkedTag(r buffer.Range) Mark {
	return Mark{Row: r.Start.Row, Span: highlight.Span{Start: r.Start.Col, End: r.End.Col, Kind: highlight.LinkedTag}}
}

// twin returns the part of the other name that matches the part from "from"
// up to "to" of the name at the cursor, if text belongs in both names.
func (e *Editor) twin(from, to buffer.Position, text string) (buffer.Range, bool) {
	name, other, ok := e.pair()
	if !ok || !contains(name, from) || !contains(name, to) || !e.doc.Type.Tags.IsName(text) {
		return buffer.Range{}, false
	}
	moved := func(p buffer.Position) buffer.Position {
		return buffer.Position{Row: other.Start.Row, Col: other.Start.Col + p.Col - name.Start.Col}
	}
	return buffer.Range{Start: moved(from), End: moved(to)}, true
}

// contains reports whether p is in r or at one of its ends.
func contains(r buffer.Range, p buffer.Position) bool {
	return !p.Before(r.Start) && !r.End.Before(p)
}

// source shows the buffer to file types, together with the highlighter
// state at the start of each line.
type source struct {
	*buffer.Buffer
	highlights *highlight.Cache
}

func (s source) State(row int) highlight.State {
	return s.highlights.State(s.Buffer, row)
}

func (e *Editor) source() filetype.Source {
	return source{e.buffer(), e.highlights}
}
