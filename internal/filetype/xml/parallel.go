package xml

import (
	"iter"
	"runtime"
	"sync"

	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/highlight"
)

// seqRows is how many rows partsFrom lexes one after the other before it
// lexes them in parallel. Most searches end within a few rows.
var seqRows = 256

// chunkRows is how many rows one goroutine lexes at most in one go. The
// chunks start smaller, so a search that ends soon after seqRows wastes
// little.
var chunkRows = 2048

const firstChunkRows = 64

// lexers is how many goroutines lex rows at the same time.
var lexers = runtime.GOMAXPROCS(0)

// partsFrom returns the parts of the rows from row on, going down for a step
// of 1 and up for -1. Going up, the parts of each row come last to first.
// After seqRows rows, goroutines lex the next batch of rows while the parts
// of the one before are used.
func partsFrom(src filetype.Source, row, step int) iter.Seq[part] {
	return func(yield func(part) bool) {
		left := func(r int) int {
			if step > 0 {
				return src.LineCount() - r
			}
			return r + 1
		}
		r := row
		var parts []part
		var buf []rune
		for n := 0; n < seqRows && left(r) > 0; n++ {
			parts, buf = appendParts(parts[:0], buf, src.Line(r), src.State(r), r, step)
			if !yieldAll(parts, yield) {
				return
			}
			r += step
		}
		size := min(firstChunkRows, chunkRows)
		next := func() func() [][]part {
			if left(r) == 0 {
				return nil
			}
			count := min(size*lexers, left(r))
			lexed := lexBatch(src, r, step, count, size)
			r += step * count
			size = min(size*2, chunkRows)
			return lexed
		}
		for lexed := next(); lexed != nil; {
			chunks := lexed()
			lexed = next()
			for _, chunk := range chunks {
				if !yieldAll(chunk, yield) {
					return
				}
			}
		}
	}
}

// lexBatch starts to lex count rows from row on in the direction of step,
// size rows per goroutine, and returns a function that waits for the parts
// of each chunk. The goroutines get the lines and states of their rows
// up front, so they read nothing that changes.
func lexBatch(src filetype.Source, row, step, count, size int) func() [][]part {
	lines := make([]string, count)
	states := make([]highlight.State, count)
	for k := range count {
		lines[k], states[k] = src.Line(row+step*k), src.State(row+step*k)
	}
	chunks := make([][]part, (count+size-1)/size)
	var wg sync.WaitGroup
	for i := range chunks {
		wg.Go(func() {
			var buf []rune
			for k := i * size; k < min((i+1)*size, count); k++ {
				chunks[i], buf = appendParts(chunks[i], buf, lines[k], states[k], row+step*k, step)
			}
		})
	}
	return func() [][]part {
		wg.Wait()
		return chunks
	}
}

func yieldAll(parts []part, yield func(part) bool) bool {
	for _, p := range parts {
		if !yield(p) {
			return false
		}
	}
	return true
}
