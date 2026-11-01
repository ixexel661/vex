package xml

import (
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/sample"
)

func BenchmarkValidate(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		for b.Loop() {
			if err := (Validator{}).Validate(doc); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkHighlightAll lexes every line, like the first ctrl+end does.
func BenchmarkHighlightAll(b *testing.B) {
	sample.Bench(b, func(b *testing.B, doc string) {
		lines := buffer.New(doc)
		for b.Loop() {
			highlight.NewCache(Highlighter{}).State(lines, lines.LineCount()-1)
		}
	})
}

func BenchmarkSchemaCheck(b *testing.B) {
	check, err := Schemas{}.Load([]config.Schema{{Path: "testdata/schemas/library.xsd"}})
	if err != nil {
		b.Fatal(err)
	}
	sample.Bench(b, func(b *testing.B, doc string) {
		for b.Loop() {
			if problems := check(doc); len(problems) != 0 {
				b.Fatal(problems[0])
			}
		}
	})
}
