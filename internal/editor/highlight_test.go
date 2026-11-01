package editor

import (
	"math/rand/v2"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/sample"
)

// TestHighlightingFollowsRandomEdits checks that after each edit the cache
// knows the same states as one that highlights the text from scratch.
func TestHighlightingFollowsRandomEdits(t *testing.T) {
	e := newEditorAt(t, xml.Type(), sample.XML(40_000)+"|")
	lines := e.buffer()
	e.highlights.State(lines, lines.LineCount()-1)
	snippets := []string{"x", "\n", "\n  <b>\n", "<!--", "-->", "<a ", "\"", "<![CDATA[", "]]>", "<?p ", "?>", ">"}
	rng := rand.New(rand.NewPCG(1, 2))
	for i := range 2000 {
		from := e.buffer().Clamp(buffer.Position{Row: rng.IntN(lines.LineCount()), Col: rng.IntN(40)})
		to := from
		if rng.IntN(3) == 0 {
			to = e.buffer().Clamp(buffer.Position{Row: from.Row + rng.IntN(4), Col: rng.IntN(40)})
		}
		text := ""
		if rng.IntN(4) > 0 {
			text = snippets[rng.IntN(len(snippets))]
		}
		e.change(buffer.Range{Start: from, End: to}, text)

		fresh := highlight.NewCache(xml.Highlighter{})
		for row := range e.highlights.Known() + 1 {
			if got, want := e.highlights.State(lines, row), fresh.State(lines, row); got != want {
				t.Fatalf("edit %d (%q at %v-%v): state of row %d = %d, want %d", i, text, from, to, row, got, want)
			}
		}
	}
	if e.highlights.Known() < lines.LineCount()/2 {
		t.Errorf("the cache knows only %d of %d rows after the edits", e.highlights.Known(), lines.LineCount())
	}
}

func TestPrefetch(t *testing.T) {
	e := newEditorAt(t, xml.Type(), "|"+sample.XML(20_000))
	job := e.Prefetch()
	if job == nil {
		t.Fatal("no job although only the first row is known")
	}
	done := job()
	e.ApplyPrefetched(done)
	if last := e.buffer().LineCount() - 1; e.highlights.Known() != last {
		t.Fatalf("knows %d rows after the job, want %d", e.highlights.Known(), last)
	}
	if e.Prefetch() != nil {
		t.Error("a new job although all rows are known")
	}

	e.highlights.Invalidate(5)
	e.InsertText("x")
	e.ApplyPrefetched(done)
	if e.highlights.Known() != 5 {
		t.Errorf("a job for the text before the edit was used, the cache knows %d rows", e.highlights.Known())
	}
}

func TestPrefetchWithoutHighlighter(t *testing.T) {
	if e := newEditor(t, "a\nb", 10, 1); e.Prefetch() != nil {
		t.Error("a job for a file type without highlighting")
	}
}
