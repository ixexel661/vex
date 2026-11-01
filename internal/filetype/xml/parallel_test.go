package xml

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/highlight"
)

var fragments = []string{
	"<a>", "</a>", "<b x=\"1\">", "</b>", "<a/>", "<b\n  y='>'>", "<c>", "</c>",
	"<!--", "-->", "<![CDATA[", "]]>", "<?p ", "?>", "<!DOCTYPE a>",
	"text", " ", "\n", "\n", "\n", "<", ">", "\"", "'", "&amp;", "</", "<a", "ü",
}

func randomDoc(rng *rand.Rand) string {
	var b strings.Builder
	for range rng.IntN(80) {
		b.WriteString(fragments[rng.IntN(len(fragments))])
	}
	return b.String()
}

// tagResults asks every method of Tags at p, all on one source, so some of
// them find the states known and others work them out.
func tagResults(text string, p buffer.Position) string {
	src := source{buffer.New(text), highlight.NewCache(Highlighter{})}
	tags := Tags{}
	name, other, paired := tags.Pair(src, p)
	end, closes := tags.Close(src, p)
	next, hasNext := tags.NextTag(src, p)
	prev, hasPrev := tags.PrevTag(src, p)
	parent, hasParent := tags.Parent(src, p)
	child, hasChild := tags.FirstChild(src, p)
	return fmt.Sprint(name, other, paired, end, closes, next, hasNext, prev, hasPrev, parent, hasParent, child, hasChild)
}

// lexingIn sets how partsFrom lexes until the test ends.
func lexingIn(t *testing.T, seq, chunk, goroutines int) {
	saved := []int{seqRows, chunkRows, lexers}
	t.Cleanup(func() { seqRows, chunkRows, lexers = saved[0], saved[1], saved[2] })
	seqRows, chunkRows, lexers = seq, chunk, goroutines
}

func TestParallelLexingFindsTheSame(t *testing.T) {
	ways := [][3]int{{0, 1, 1}, {0, 2, 3}, {1, 3, 2}, {2, 1, 4}}
	rng := rand.New(rand.NewPCG(1, 2))
	for range 500 {
		text := randomDoc(rng)
		lines := strings.Split(text, "\n")
		for range 8 {
			row := rng.IntN(len(lines))
			p := buffer.Position{Row: row, Col: rng.IntN(utf8.RuneCountInString(lines[row]) + 1)}
			lexingIn(t, 1<<30, chunkRows, lexers)
			want := tagResults(text, p)
			for _, way := range ways {
				lexingIn(t, way[0], way[1], way[2])
				if got := tagResults(text, p); got != want {
					t.Fatalf("lexing in %v at %v of %q:\n got %s\nwant %s", way, p, text, got, want)
				}
			}
		}
	}
}

func TestLexPartsFindsTheParts(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	for range 2000 {
		line := []rune(strings.ReplaceAll(randomDoc(rng), "\n", ""))
		for state := inText; state <= inDeclaration; state++ {
			want := lex(line, state).parts
			if got := lexParts(nil, line, state); !slices.Equal(got, want) {
				t.Fatalf("lexParts(%q, %d) = %v, want %v", string(line), state, got, want)
			}
		}
	}
}

func TestIsNameRune(t *testing.T) {
	for r := range rune(unicode.MaxRune + 1) {
		want := unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_:-.", r)
		if got := isNameRune(r); got != want {
			t.Fatalf("isNameRune(%q) = %v, want %v", r, got, want)
		}
	}
}
