package buffer

import (
	"strings"
	"sync"
	"testing"
)

func TestSnapshotKeepsItsText(t *testing.T) {
	edits := map[string]func(b *Buffer){
		"insert":  func(b *Buffer) { b.Insert(Position{Row: 1, Col: 1}, "x") },
		"newline": func(b *Buffer) { b.Insert(Position{Row: 0, Col: 1}, "\n") },
		"delete":  func(b *Buffer) { b.Delete(Position{Row: 0, Col: 1}, Position{Row: 2, Col: 0}) },
		"replace": func(b *Buffer) { b.SetText("x") },
	}
	for name, edit := range edits {
		b := New("ab\ncd\nef")
		s := b.Snapshot()
		edit(b)
		if got := s.Text(); got != "ab\ncd\nef" {
			t.Errorf("%s: snapshot text = %q, want the text from before", name, got)
		}
		s.Release()
	}
}

// TestReleasedSnapshotNeedsNoCopy checks whether an edit that keeps the
// number of lines changes the list of lines in place.
func TestReleasedSnapshotNeedsNoCopy(t *testing.T) {
	b := New("ab\ncd")
	inPlace := func() bool {
		first := &b.lines[0]
		b.Insert(Position{Row: 1, Col: 0}, "x")
		return first == &b.lines[0]
	}
	if !inPlace() {
		t.Fatal("an edit without snapshots copied the lines")
	}
	s := b.Snapshot()
	if inPlace() {
		t.Error("an edit changed the lines a snapshot reads")
	}
	if !inPlace() {
		t.Error("the second edit after a snapshot copied the lines again")
	}
	s.Release()
	if !inPlace() {
		t.Error("an edit after the release copied the lines")
	}

	b.Snapshot()
	b.SetText("ab\ncd")
	if !inPlace() {
		t.Error("an edit after SetText copied the new lines")
	}
}

func TestReleaseCountsOnce(t *testing.T) {
	b := New("ab")
	first, second := b.Snapshot(), b.Snapshot()
	first.Release()
	first.Release()
	b.Insert(Position{}, "x")
	if got := second.Text(); got != "ab" {
		t.Errorf("snapshot text = %q after the other one was released twice, want %q", got, "ab")
	}
	second.Release()
}

func TestSnapshotsCanBeReadWhileTheBufferChanges(t *testing.T) {
	b := New(strings.Repeat("line\n", 1000))
	var wg sync.WaitGroup
	for range 4 {
		s := b.Snapshot()
		want := s.Text()
		wg.Go(func() {
			defer s.Release()
			for range 50 {
				if s.Text() != want {
					t.Error("a snapshot changed with the buffer")
					return
				}
			}
		})
		for row := range 100 {
			b.Insert(Position{Row: row}, "x")
			b.Delete(Position{Row: row}, Position{Row: row, Col: 1})
		}
	}
	wg.Wait()
}
