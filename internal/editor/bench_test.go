package editor

import (
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/sample"
	"github.com/ixexel661/vex/internal/theme"
)

func benchEditor(doc string) *Editor {
	d := &document.Document{Path: "bench.xml", Type: xml.Type(), Buffer: buffer.New(doc)}
	e := New(d, config.Default(), theme.Theme{})
	e.SetSize(120, 40)
	return e
}

// rootEnd returns where the name of the root's end tag ends.
func rootEnd(e *Editor) buffer.Position {
	last := e.buffer().LineCount() - 2 // the last line is empty
	return buffer.Position{Row: last, Col: len("</library")}
}

// BenchmarkRenameRootAtTheEnd types in the end tag of the root element at
// the bottom, so linked editing also changes the start tag at the top.
func BenchmarkRenameRootAtTheEnd(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		e.SetLinkedEditing(true)
		e.MoveTo(rootEnd(e))
		e.ApplyFoundPair(e.FindPair()())
		e.View()
		for b.Loop() {
			e.InsertText("x")
			e.View()
		}
		if !strings.HasPrefix(e.buffer().Line(1), "<libraryx") {
			b.Fatal("the start tag was not renamed")
		}
	})
}

// BenchmarkMoveOntoTheRootEnd moves onto the end tag of the root, which
// starts a job that looks for its start tag.
func BenchmarkMoveOntoTheRootEnd(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		e.SetLinkedEditing(true)
		end := rootEnd(e)
		e.MoveTo(end)
		e.View()
		for b.Loop() {
			e.knownPair = nil
			e.MoveTo(end)
			e.View()
			if e.FindPair() == nil {
				b.Fatal("no job looks for the start tag")
			}
		}
	})
}

// BenchmarkMoveToRootFromTheEnd presses ctrl+up on the end tag of the root.
func BenchmarkMoveToRootFromTheEnd(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		end := buffer.Position{Row: rootEnd(e).Row}
		e.MoveTo(end)
		e.View()
		for b.Loop() {
			e.MoveTo(end)
			e.MoveToParent()
		}
	})
}

// BenchmarkFindRootPair is the job that looks for the start tag of the root
// from its end tag.
func BenchmarkFindRootPair(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		e.SetLinkedEditing(true)
		e.MoveTo(rootEnd(e))
		job := e.FindPair()
		for b.Loop() {
			if !job().ok {
				b.Fatal("the job did not find the start tag")
			}
		}
	})
}

// BenchmarkEditThenGoToEnd types at the top and then presses ctrl+end.
func BenchmarkEditThenGoToEnd(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		e.MoveBottom()
		e.View()
		for b.Loop() {
			e.MoveTop()
			e.InsertText("x")
			e.MoveBottom()
			e.View()
		}
	})
}

// BenchmarkCloseTagUnderTheRoot types a new start tag at the top of the root,
// so ">" looks for an end tag down to the end of the root.
func BenchmarkCloseTagUnderTheRoot(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		e.MoveBottom()
		e.View()
		at := buffer.Position{Row: 2}
		for b.Loop() {
			e.MoveTo(at)
			e.InsertText("<x")
			e.CloseTag()
			if got := e.buffer().Line(2); !strings.HasPrefix(got, "<x></x>") {
				b.Fatalf("line = %q, want the end tag added", got)
			}
			e.MoveTo(at)
			for range len("<x></x>") {
				e.DeleteForward()
			}
		}
	})
}

// BenchmarkFirstChildFromTheEnd presses ctrl+down on the comment before the
// last book, which goes to the first child of the root.
func BenchmarkFirstChildFromTheEnd(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		e := benchEditor(doc)
		e.MoveBottom()
		e.View()
		comment := buffer.Position{Row: e.buffer().LineCount() - 13}
		first := buffer.Position{Row: 3, Col: 2}
		for b.Loop() {
			e.MoveTo(comment)
			e.MoveToFirstChild()
			if e.Cursor() != first {
				b.Fatalf("cursor = %v, want the first book at %v", e.Cursor(), first)
			}
		}
	})
}
