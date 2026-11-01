package search

import (
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/sample"
)

func BenchmarkTextFind(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		lines := buffer.New(doc)
		for b.Loop() {
			Text{}.Find(lines, "author")
		}
	})
}

func BenchmarkFuzzyFind(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		lines := buffer.New(doc)
		for b.Loop() {
			Fuzzy{}.Find(lines, "athr")
		}
	})
}
