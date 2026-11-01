package buffer

import (
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

// Snapshot is the text of a buffer at one moment. Other goroutines can read
// it while the buffer goes on changing, until they release it.
type Snapshot struct {
	lines   []string
	release func()
}

// Snapshot shares the list of lines with the buffer. The buffer copies the
// list before it changes it, as long as a snapshot is not released.
func (b *Buffer) Snapshot() Snapshot {
	if b.readers == nil {
		b.readers = new(atomic.Int32)
	}
	readers := b.readers
	readers.Add(1)
	return Snapshot{lines: b.lines, release: sync.OnceFunc(func() { readers.Add(-1) })}
}

// Release tells the buffer that the snapshot is no longer read, so the buffer
// can change its lines in place again.
func (s Snapshot) Release() {
	s.release()
}

// own copies the lines before the buffer changes them, if a snapshot still
// reads them.
func (b *Buffer) own() {
	if b.readers != nil && b.readers.Load() > 0 {
		b.lines = slices.Clone(b.lines)
		b.readers = nil
	}
}

func (s Snapshot) Line(row int) string {
	return s.lines[row]
}

func (s Snapshot) LineCount() int {
	return len(s.lines)
}

func (s Snapshot) Text() string {
	return strings.Join(s.lines, "\n")
}
