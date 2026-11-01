package editor

import "github.com/ixexel661/vex/internal/highlight"

// Prefetched is highlighting that was worked out in the background.
type Prefetched struct {
	version int
	row     int
	states  []highlight.State
}

// Prefetch returns a job that highlights the rows below the ones known so
// far, e.g. for ctrl+end. It works on a snapshot of the text, so it can run
// in another goroutine. Without such rows, it returns nil.
func (e *Editor) Prefetch() func() Prefetched {
	h := e.doc.Type.Highlighter
	row := e.highlights.Known()
	if h == nil || row >= e.buffer().LineCount()-1 {
		return nil
	}
	version, lines, start := e.buffer().Version(), e.buffer().Snapshot(), e.highlights.State(e.buffer(), row)
	return func() Prefetched {
		defer lines.Release()
		return Prefetched{version: version, row: row, states: highlight.States(h, lines, row, start)}
	}
}

// ApplyPrefetched keeps what a job of Prefetch worked out, unless the text
// changed since.
func (e *Editor) ApplyPrefetched(p Prefetched) {
	if p.version == e.buffer().Version() {
		e.highlights.Extend(p.row, p.states)
	}
}
