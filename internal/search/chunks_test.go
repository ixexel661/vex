package search

import (
	"reflect"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/sample"
)

func TestInChunksKeepsTheOrder(t *testing.T) {
	defer func(n int) { chunkLines = n }(chunkLines)
	chunkLines = 3
	for _, count := range []int{0, 1, 3, 7, 9, 100} {
		got := inChunks(count, func(from, to int) []int {
			var rows []int
			for row := from; row < to; row++ {
				rows = append(rows, row)
			}
			return rows
		})
		if len(got) != count {
			t.Fatalf("%d rows: got %d", count, len(got))
		}
		for i, row := range got {
			if row != i {
				t.Fatalf("%d rows: got %v", count, got)
			}
		}
	}
}

// TestParallelSearchFindsTheSame compares the search in pieces with the one
// in a single piece.
func TestParallelSearchFindsTheSame(t *testing.T) {
	lines := buffer.New(sample.XML(5 << 20))
	for _, m := range []Matcher{Text{}, Fuzzy{}} {
		for _, query := range []string{"author", "Autor 7", "<tag>", "über", "Über"} {
			parallel := m.Find(lines, query)
			func() {
				defer func(n int) { chunkLines = n }(chunkLines)
				chunkLines = lines.LineCount()
				if sequential := m.Find(lines, query); !reflect.DeepEqual(parallel, sequential) {
					t.Errorf("%s %q: %d matches in pieces, %d in one go, or in another order", m.Name(), query, len(parallel), len(sequential))
				}
			}()
			if len(parallel) == 0 {
				t.Errorf("%s %q found nothing", m.Name(), query)
			}
		}
	}
}
