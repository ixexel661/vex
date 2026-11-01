package editor

import (
	"runtime"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype/xml"
)

// copiesLines reports whether typing copies the list of lines, which the
// buffer does while a snapshot is not released.
func copiesLines(e *Editor) bool {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	e.MoveTo(buffer.Position{Row: 1})
	e.InsertText(" ")
	runtime.ReadMemStats(&after)
	return after.TotalAlloc-before.TotalAlloc > uint64(e.buffer().LineCount()*8)
}

// run turns a job into a function without a result, or nil without a job.
func run[T any](job func() T) func() {
	if job == nil {
		return nil
	}
	return func() { job() }
}

func TestJobsReleaseTheirSnapshots(t *testing.T) {
	jobs := map[string]func(e *Editor) func(){
		"pair": func(e *Editor) func() {
			e.SetLinkedEditing(true)
			return run(e.FindPair())
		},
		"jump": func(e *Editor) func() {
			e.MoveToParent()
			return run(e.FindJump())
		},
		"prefetch": func(e *Editor) func() {
			return run(e.Prefetch())
		},
	}
	// The cursor is in the end tag of the root, whose start tag is far.
	for name, start := range jobs {
		e := newEditorAt(t, xml.Type(), apart(nearRows+1, true))
		job := start(e)
		if job == nil {
			t.Fatalf("%s: no job", name)
		}
		if !copiesLines(e) {
			t.Fatalf("%s: typing does not copy the lines while the job reads them", name)
		}
		e = newEditorAt(t, xml.Type(), apart(nearRows+1, true))
		start(e)()
		if copiesLines(e) {
			t.Errorf("%s: typing copies the lines after the job is done", name)
		}
	}
}
