package editor

import (
	"unicode/utf8"

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

func (e *Editor) MoveToNextTag() {
	e.jump(filetype.Tags.NextTag)
}

func (e *Editor) MoveToPrevTag() {
	e.jump(filetype.Tags.PrevTag)
}

func (e *Editor) MoveToParent() {
	e.jumpNearFirst(filetype.Tags.Parent)
}

func (e *Editor) MoveToFirstChild() {
	e.jump(filetype.Tags.FirstChild)
}

// jump moves the cursor to the tag that find returns, if the file type has
// tags.
func (e *Editor) jump(find func(filetype.Tags, filetype.Source, buffer.Position) (buffer.Position, bool)) {
	tags := e.doc.Type.Tags
	if tags == nil {
		return
	}
	if p, ok := find(tags, e.source(), e.cursor); ok {
		e.setCursor(p)
	}
}

// nearRows is how far from the cursor the editor looks for the other tag of
// a pair at once. Further away, a job of FindPair looks for it.
const nearRows = 5000

// tagPair is a tag name and what the editor knows about the other name of its
// element, in one version of the text.
type tagPair struct {
	version     int
	name, other buffer.Range
	ok          bool // other is the other name
	far         bool // the other name may be further away than nearRows
}

// pair returns the tag name at the cursor and the other name of its element.
func (e *Editor) pair() (name, other buffer.Range, ok bool) {
	p := e.cursorPair()
	if p == nil {
		return name, other, false
	}
	return p.name, p.other, p.ok
}

// cursorPair returns the tag pair at the cursor, or nil if the cursor is not
// at a tag name. It looks for each name only once per version of the text.
func (e *Editor) cursorPair() *tagPair {
	tags := e.doc.Type.Tags
	if !e.linkedEditing || tags == nil {
		return nil
	}
	version := e.buffer().Version()
	if p := e.knownPair; p != nil && p.version == version && contains(p.name, e.cursor) {
		return p
	}
	w := e.near()
	name, other, ok := tags.Pair(w, w.in(e.cursor))
	if name == (buffer.Range{}) {
		return nil
	}
	e.knownPair = &tagPair{version: version, name: w.outRange(name), other: w.outRange(other), ok: ok, far: !ok && w.cut()}
	return e.knownPair
}

// renamed returns the pair after the same change went into both names: the
// text from "from" up to "to" of the name became text.
func (p *tagPair) renamed(from, to buffer.Position, text string, version int) *tagPair {
	delta := utf8.RuneCountInString(text) - (to.Col - from.Col)
	name, other := p.name, p.other
	name.End.Col += delta
	other.End.Col += delta
	if name.Start.Row == other.Start.Row {
		later := &other
		if other.Start.Before(name.Start) {
			later = &name
		}
		later.Start.Col += delta
		later.End.Col += delta
	}
	return &tagPair{version: version, name: name, other: other, ok: true}
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

// source shows the text to file types, together with the highlighter state
// at the start of each line.
type source struct {
	highlight.Lines
	highlights *highlight.Cache
}

func (s source) State(row int) highlight.State {
	return s.highlights.State(s.Lines, row)
}

func (e *Editor) source() filetype.Source {
	return source{e.buffer(), e.highlights}
}

// snapshot returns the text and its highlighting for another goroutine,
// which calls release when it is done with them.
func (e *Editor) snapshot() (src filetype.Source, release func()) {
	lines := e.buffer().Snapshot()
	return source{lines, e.highlights.Clone()}, lines.Release
}

// window shows the rows from top up to bottom of a source as a document of
// their own, so a search in it stops at its edges.
type window struct {
	filetype.Source
	top, bottom int
}

// near returns the rows up to nearRows away from the cursor.
func (e *Editor) near() window {
	return window{
		Source: e.source(),
		top:    max(e.cursor.Row-nearRows, 0),
		bottom: min(e.cursor.Row+nearRows+1, e.buffer().LineCount()),
	}
}

func (w window) LineCount() int {
	return w.bottom - w.top
}

func (w window) Line(row int) string {
	return w.Source.Line(w.top + row)
}

func (w window) State(row int) highlight.State {
	return w.Source.State(w.top + row)
}

// cut reports whether the window leaves out rows of its source.
func (w window) cut() bool {
	return w.top > 0 || w.bottom < w.Source.LineCount()
}

// in turns a position in the source into one in the window, out and
// outRange do the opposite.
func (w window) in(p buffer.Position) buffer.Position {
	p.Row -= w.top
	return p
}

func (w window) out(p buffer.Position) buffer.Position {
	p.Row += w.top
	return p
}

func (w window) outRange(r buffer.Range) buffer.Range {
	return buffer.Range{Start: w.out(r.Start), End: w.out(r.End)}
}
