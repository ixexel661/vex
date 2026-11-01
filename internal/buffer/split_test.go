package buffer

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/sample"
)

func randomText(rng *rand.Rand) string {
	parts := []string{"a", "b", "ü", "\n", "\r", "\r\n"}
	var b strings.Builder
	for range rng.IntN(60) {
		b.WriteString(parts[rng.IntN(len(parts))])
	}
	return b.String()
}

// TestSplitInPieces compares split, with tiny pieces, with strings.Split.
func TestSplitInPieces(t *testing.T) {
	defer func(n int) { splitBytes = n }(splitBytes)
	rng := rand.New(rand.NewPCG(3, 4))
	for range 5000 {
		text := randomText(rng)
		splitBytes = 1 + rng.IntN(8)
		if got, want := split(text, false), strings.Split(text, "\n"); !slices.Equal(got, want) {
			t.Fatalf("split(%q) in pieces of %d = %q, want %q", text, splitBytes, got, want)
		}
		if got, want := split(text, true), strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n"); !slices.Equal(got, want) {
			t.Fatalf("split(%q) without CR in pieces of %d = %q, want %q", text, splitBytes, got, want)
		}
	}
}

func TestNewSplitsLargeTexts(t *testing.T) {
	text := sample.XML(5 << 20)
	if got := New(text).Text(); got != text {
		t.Error("the text of a large buffer differs from the text it was made of")
	}
	crlf := strings.ReplaceAll(text, "\n", "\r\n")
	if got := NewCRLF(crlf).Text(); got != text {
		t.Error("the text of a large CRLF buffer still has its \\r")
	}
}
