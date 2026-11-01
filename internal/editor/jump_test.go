package editor

import (
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype/xml"
)

// underRoot returns a root with rows children, with the cursor at the last
// one, rows rows below the start tag of the root.
func underRoot(rows int) string {
	return "<a>\n" + strings.Repeat("<b/>\n", rows-1) + "|<b/>\n</a>"
}

func TestNearParentJumpsAtOnce(t *testing.T) {
	e := newEditorAt(t, xml.Type(), underRoot(nearRows))
	e.MoveToParent()
	if got := e.Cursor(); got != (buffer.Position{}) {
		t.Errorf("cursor = %v, want the start tag of the root %d rows up", got, nearRows)
	}
	if e.Jumping() || e.FindJump() != nil {
		t.Error("a job looks for a parent found near the cursor")
	}
}

func TestFarParentIsFoundByAJob(t *testing.T) {
	e := newEditorAt(t, xml.Type(), underRoot(nearRows+1))
	from := e.Cursor()
	e.MoveToParent()
	if e.Cursor() != from || !e.Jumping() {
		t.Fatalf("cursor = %v, jumping = %v, want it waiting at %v", e.Cursor(), e.Jumping(), from)
	}
	job := e.FindJump()
	if job == nil {
		t.Fatal("no job looks for the far parent")
	}
	e.ApplyJump(job())
	if got := e.Cursor(); got != (buffer.Position{}) {
		t.Errorf("cursor = %v, want the start tag of the root", got)
	}
	if e.Jumping() || e.FindJump() != nil {
		t.Error("the jump still waits after it was made")
	}
}

func TestFarJumpIsCancelled(t *testing.T) {
	for name, cancel := range map[string]func(*Editor){
		"move right":         (*Editor).MoveRight,
		"move down":          (*Editor).MoveDown,
		"move away and back": func(e *Editor) { e.MoveRight(); e.MoveLeft() },
		"type":               func(e *Editor) { e.InsertText("x") },
	} {
		e := newEditorAt(t, xml.Type(), underRoot(nearRows+1))
		e.MoveToParent()
		job := e.FindJump()
		cancel(e)
		before := e.Cursor()
		e.ApplyJump(job())
		if e.Cursor() != before || e.Jumping() {
			t.Errorf("%s: cursor = %v, jumping = %v, want it to stay at %v", name, e.Cursor(), e.Jumping(), before)
		}
	}
}

func TestOldJumpIsDropped(t *testing.T) {
	e := newEditorAt(t, xml.Type(), underRoot(nearRows+2))
	e.MoveToParent()
	old := e.FindJump()
	e.MoveUp()
	e.MoveToParent()
	waiting := e.Cursor()
	e.ApplyJump(old())
	if e.Cursor() != waiting || !e.Jumping() {
		t.Errorf("cursor = %v, jumping = %v, want the newer jump still waiting at %v", e.Cursor(), e.Jumping(), waiting)
	}
}

func TestNoJumpJobWhenTheWholeTextWasSearched(t *testing.T) {
	e := newEditorAt(t, xml.Type(), "<a/>|")
	e.MoveToParent()
	if e.Jumping() || e.FindJump() != nil {
		t.Error("a job looks for a parent in a text that was searched completely")
	}
}

func TestFindJumpWorksOnASnapshot(t *testing.T) {
	e := newEditorAt(t, xml.Type(), underRoot(nearRows+1))
	e.MoveToParent()
	job := e.FindJump()
	e.MoveTo(buffer.Position{})
	e.InsertText("<!--")
	if !job().ok {
		t.Error("the job saw the text commented out after it started")
	}
}
