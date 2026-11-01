package buffer

import (
	"testing"

	"github.com/ixexel661/vex/internal/sample"
)

func BenchmarkNew(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		for b.Loop() {
			New(doc)
		}
	})
}
