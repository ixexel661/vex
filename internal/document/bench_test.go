package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/sample"
)

func BenchmarkOpen(b *testing.B) {
	for _, ending := range []string{"LF", "CRLF"} {
		b.Run(ending, func(b *testing.B) {
			sample.Bench(b, func(b *testing.B, doc string) {
				if ending == "CRLF" {
					doc = strings.ReplaceAll(doc, "\n", "\r\n")
				}
				path := filepath.Join(b.TempDir(), "doc.xml")
				if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
					b.Fatal(err)
				}
				for b.Loop() {
					if _, err := Open(path, filetype.Plain()); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
