package editor

import (
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/filetype/xml"
)

// apart returns an element whose end tag is rows rows below its start tag,
// with the cursor in the name of its start or end tag.
func apart(rows int, inEndTag bool) string {
	start, end := "<a|>", "</a>"
	if inEndTag {
		start, end = "<a>", "</a|>"
	}
	return start + "\n" + strings.Repeat("<b/>\n", rows-1) + end
}

func TestNearPairNeedsNoJob(t *testing.T) {
	for _, inEndTag := range []bool{false, true} {
		e := newTagEditor(t, xml.Type(), apart(nearRows, inEndTag))
		if _, _, ok := e.pair(); !ok {
			t.Errorf("no pair %d rows apart, cursor in the end tag: %v", nearRows, inEndTag)
		}
		if e.FindPair() != nil {
			t.Errorf("a job looks for a pair found near the cursor, cursor in the end tag: %v", inEndTag)
		}
	}
}

func TestNoJobWithoutAPairInASmallText(t *testing.T) {
	e := newTagEditor(t, xml.Type(), "<a|>\n<b/>")
	if _, _, ok := e.pair(); ok {
		t.Fatal("found a pair for a tag without an end tag")
	}
	if e.FindPair() != nil {
		t.Error("a job looks for a pair in a text that was searched completely")
	}
}

func TestFarPairIsFoundByAJob(t *testing.T) {
	for _, inEndTag := range []bool{false, true} {
		e := newTagEditor(t, xml.Type(), apart(nearRows+1, inEndTag))
		if _, _, ok := e.pair(); ok {
			t.Fatalf("the pair %d rows apart was looked for in the whole text", nearRows+1)
		}
		job := e.FindPair()
		if job == nil {
			t.Fatalf("no job looks for the far pair, cursor in the end tag: %v", inEndTag)
		}
		e.ApplyFoundPair(job())
		if _, _, ok := e.pair(); !ok {
			t.Fatalf("the job did not find the pair, cursor in the end tag: %v", inEndTag)
		}
		if e.FindPair() != nil {
			t.Error("a job looks for a known pair")
		}

		e.InsertText("x")
		e.InsertText("y")
		if text := e.Document().Buffer.Text(); !strings.HasPrefix(text, "<axy>") || !strings.HasSuffix(text, "</axy>") {
			t.Errorf("typing in a far pair renamed %q … %q", text[:5], text[len(text)-6:])
		}
	}
}

func TestFindPairOnlyWithLinkedEditing(t *testing.T) {
	e := newTagEditor(t, xml.Type(), apart(nearRows+1, true))
	e.SetLinkedEditing(false)
	if e.FindPair() != nil {
		t.Error("a job looks for a pair without linked editing")
	}
}

func TestFoundPairOfOldTextIsDropped(t *testing.T) {
	e := newTagEditor(t, xml.Type(), apart(nearRows+1, true))
	at := e.Cursor()
	job := e.FindPair()
	e.MoveTo(buffer.Position{Row: at.Row - 1})
	e.InsertText(" ")
	e.MoveTo(at)
	e.ApplyFoundPair(job())
	if _, _, ok := e.pair(); ok {
		t.Error("the pair found in the old text was kept")
	}
	if e.FindPair() == nil {
		t.Error("no job looks for the pair in the new text")
	}
}

func TestFindPairWorksOnASnapshot(t *testing.T) {
	e := newTagEditor(t, xml.Type(), apart(nearRows+1, true))
	job := e.FindPair()
	e.MoveTo(buffer.Position{})
	e.InsertText("<!--")
	if !job().ok {
		t.Error("the job saw the text commented out after it started")
	}
}

// countingTags counts how often the editor looks for tag pairs.
type countingTags struct {
	xml.Tags
	pairs int
}

func (c *countingTags) Pair(src filetype.Source, p buffer.Position) (name, other buffer.Range, ok bool) {
	c.pairs++
	return c.Tags.Pair(src, p)
}

func TestPairIsLookedForOncePerVersion(t *testing.T) {
	tags := &countingTags{}
	fileType := xml.Type()
	fileType.Tags = tags
	e := newTagEditor(t, fileType, "<bo|ok>x</book>")
	e.View()
	e.MoveRight()
	e.View()
	e.InsertText("x")
	e.View()
	if tags.pairs != 1 {
		t.Errorf("looked for the pair %d times, want once", tags.pairs)
	}
	e.MoveLineEnd()
	e.InsertText(" ")
	e.MoveTo(buffer.Position{Col: 1})
	tags.pairs = 0
	e.View()
	e.View()
	if tags.pairs != 1 {
		t.Errorf("looked for the pair %d times after an edit outside the names, want once", tags.pairs)
	}
}
