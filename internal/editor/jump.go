package editor

import (
	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
)

type findTag func(filetype.Tags, filetype.Source, buffer.Position) (buffer.Position, bool)

// farJump is a jump to a tag further away than the search near the cursor
// went. Moving the cursor cancels it.
type farJump struct {
	version int
	from    buffer.Position
	find    findTag
}

// Jump is where a job of FindJump found the tag to jump to.
type Jump struct {
	version  int
	from, to buffer.Position
	ok       bool
}

// jumpNearFirst is jump for searches that go one way from the cursor, so a
// tag they find near the cursor is the one they find in the whole text.
// Further away, a job of FindJump looks for it.
func (e *Editor) jumpNearFirst(find findTag) {
	tags := e.doc.Type.Tags
	if tags == nil {
		return
	}
	w := e.near()
	if p, ok := find(tags, w, w.in(e.cursor)); ok {
		e.setCursor(w.out(p))
		return
	}
	if w.cut() {
		e.farJump = &farJump{version: e.buffer().Version(), from: e.cursor, find: find}
	}
}

// Jumping reports whether the cursor waits for a job of FindJump.
func (e *Editor) Jumping() bool {
	return e.farJump != nil && e.farJump.version == e.buffer().Version()
}

// FindJump returns a job that looks for the tag the cursor waits for in the
// whole text. It works on a snapshot, so it can run in another goroutine.
// Without a waiting jump, it returns nil.
func (e *Editor) FindJump() func() Jump {
	if !e.Jumping() {
		return nil
	}
	j, tags := *e.farJump, e.doc.Type.Tags
	src, release := e.snapshot()
	return func() Jump {
		defer release()
		to, ok := j.find(tags, src, j.from)
		return Jump{version: j.version, from: j.from, to: to, ok: ok}
	}
}

// ApplyJump moves the cursor to what a job of FindJump found, unless the
// cursor or the text changed since.
func (e *Editor) ApplyJump(j Jump) {
	if !e.Jumping() || e.farJump.version != j.version || e.farJump.from != j.from {
		return
	}
	e.farJump = nil
	if j.ok {
		e.setCursor(j.to)
	}
}
