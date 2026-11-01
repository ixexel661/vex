package buffer

import (
	"strings"
	"sync"
)

// splitBytes is about how much text one goroutine splits into lines.
// Shorter texts are split in one go.
var splitBytes = 1 << 20

// split cuts text into lines at "\n". With dropCR, a "\r" right before the
// "\n" goes too. Long texts are cut at line breaks into pieces first, which
// are counted and split at the same time.
func split(text string, dropCR bool) []string {
	pieces := cut(text)
	starts := make([]int, len(pieces)+1) // the first line of each piece
	inParallel(len(pieces), func(i int) {
		starts[i+1] = strings.Count(pieces[i], "\n") + 1
	})
	for i := range pieces {
		starts[i+1] += starts[i]
	}
	lines := make([]string, starts[len(pieces)])
	inParallel(len(pieces), func(i int) {
		last := i == len(pieces)-1
		splitInto(lines[starts[i]:starts[i+1]], pieces[i], dropCR, !last)
	})
	return lines
}

// cut cuts text into pieces of about splitBytes at line breaks, which are
// left out.
func cut(text string) []string {
	var pieces []string
	for len(text) > splitBytes {
		i := strings.IndexByte(text[splitBytes:], '\n')
		if i < 0 {
			break
		}
		pieces = append(pieces, text[:splitBytes+i])
		text = text[splitBytes+i+1:]
	}
	return append(pieces, text)
}

// splitInto puts the lines of s into dst, which has room for exactly them.
// The last one only ends in a line break if more text follows.
func splitInto(dst []string, s string, dropCR, lastEndsLine bool) {
	for i := range dst {
		line := s
		if i < len(dst)-1 {
			end := strings.IndexByte(s, '\n')
			line, s = s[:end], s[end+1:]
		}
		if dropCR && (i < len(dst)-1 || lastEndsLine) {
			line = strings.TrimSuffix(line, "\r")
		}
		dst[i] = line
	}
}

func inParallel(n int, do func(i int)) {
	if n == 1 {
		do(0)
		return
	}
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() { do(i) })
	}
	wg.Wait()
}
