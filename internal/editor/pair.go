package editor

import "github.com/ixexel661/vex/internal/buffer"

// FoundPair is a tag pair that a job of FindPair looked for in the whole text.
type FoundPair struct {
	version     int
	name, other buffer.Range
	ok          bool
}

// FindPair returns a job that looks for the other name of the tag at the
// cursor in the whole text, if it is not near the cursor. It works on a
// snapshot of the text, so it can run in another goroutine. Without such a
// tag, it returns nil.
func (e *Editor) FindPair() func() FoundPair {
	p := e.cursorPair()
	if p == nil || !p.far {
		return nil
	}
	tags, version, at := e.doc.Type.Tags, p.version, p.name.Start
	src, release := e.snapshot()
	return func() FoundPair {
		defer release()
		name, other, ok := tags.Pair(src, at)
		return FoundPair{version: version, name: name, other: other, ok: ok}
	}
}

// ApplyFoundPair keeps what a job of FindPair found, unless the text changed
// since.
func (e *Editor) ApplyFoundPair(f FoundPair) {
	if f.version == e.buffer().Version() {
		e.knownPair = &tagPair{version: f.version, name: f.name, other: f.other, ok: f.ok}
	}
}
