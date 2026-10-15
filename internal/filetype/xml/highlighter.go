package xml

import (
	"slices"
	"strings"
	"unicode"

	"github.com/ixexel661/vex/internal/highlight"
)

const (
	inText highlight.State = iota
	inTag
	inDoubleQuoted
	inSingleQuoted
	inComment
	inCData
	inProcInst
	inDeclaration
)

// block is a construct that runs until its closing delimiter, which may be
// several lines further down.
type block struct {
	open, close string
	kind        highlight.Kind
	after       highlight.State
}

var blocks = map[highlight.State]block{
	inDoubleQuoted: {`"`, `"`, highlight.String, inTag},
	inSingleQuoted: {"'", "'", highlight.String, inTag},
	inComment:      {"<!--", "-->", highlight.Comment, inText},
	inCData:        {"<![CDATA[", "]]>", highlight.CData, inText},
	inProcInst:     {"<?", "?>", highlight.Meta, inText},
	inDeclaration:  {"<!", ">", highlight.Meta, inText},
}

// Comments and CDATA come first, because "<!" also matches their openers.
var markup = []highlight.State{inComment, inCData, inProcInst, inDeclaration}

// Highlighter is tolerant of broken XML, which is common while typing.
type Highlighter struct{}

func (Highlighter) Highlight(line []rune, state highlight.State) ([]highlight.Span, highlight.State) {
	l := &lexer{line: line, state: state}
	for l.pos < len(l.line) {
		l.step()
	}
	return l.spans, l.state
}

type lexer struct {
	line  []rune
	pos   int
	state highlight.State
	spans []highlight.Span
}

func (l *lexer) step() {
	switch l.state {
	case inText:
		l.text()
	case inTag:
		l.tag()
	default:
		l.block(l.state, 0)
	}
}

func (l *lexer) text() {
	for _, state := range markup {
		if open := blocks[state].open; l.hasPrefix(open) {
			l.block(state, len(open))
			return
		}
	}
	switch l.line[l.pos] {
	case '<':
		l.openTag()
	case '&':
		l.entity()
	default:
		l.takeWhile(highlight.Plain, func(r rune) bool { return r != '<' && r != '&' })
	}
}

func (l *lexer) openTag() {
	if l.hasPrefix("</") {
		l.take(2, highlight.Punctuation)
	} else {
		l.take(1, highlight.Punctuation)
	}
	l.takeWhile(highlight.Tag, isNameRune)
	l.state = inTag
}

func (l *lexer) tag() {
	r := l.line[l.pos]
	switch {
	case l.hasPrefix("/>"):
		l.take(2, highlight.Punctuation)
		l.state = inText
	case r == '>':
		l.take(1, highlight.Punctuation)
		l.state = inText
	case r == '<':
		// The tag was never closed, a new one starts here.
		l.state = inText
	case r == '=':
		l.take(1, highlight.Punctuation)
	case r == '"':
		l.block(inDoubleQuoted, 1)
	case r == '\'':
		l.block(inSingleQuoted, 1)
	case isNameRune(r):
		l.takeWhile(highlight.Attribute, isNameRune)
	default:
		l.take(1, highlight.Plain)
	}
}

func (l *lexer) entity() {
	n := 1
	for l.pos+n < len(l.line) && (isNameRune(l.line[l.pos+n]) || l.line[l.pos+n] == '#') {
		n++
	}
	if l.pos+n < len(l.line) && l.line[l.pos+n] == ';' {
		l.take(n+1, highlight.Entity)
		return
	}
	l.take(1, highlight.Plain)
}

// block highlights from the current position, skipping the opener, up to and
// including the closing delimiter. If the line ends first, the block
// continues on the next line, which the state remembers.
func (l *lexer) block(state highlight.State, skip int) {
	b := blocks[state]
	l.state = state
	end := len(l.line)
	if i := index(l.line[l.pos+skip:], b.close); i >= 0 {
		end = l.pos + skip + i + len(b.close)
		l.state = b.after
	}
	l.take(end-l.pos, b.kind)
}

func (l *lexer) take(n int, kind highlight.Kind) {
	l.spans = append(l.spans, highlight.Span{Start: l.pos, End: l.pos + n, Kind: kind})
	l.pos += n
}

func (l *lexer) takeWhile(kind highlight.Kind, match func(rune) bool) {
	n := 0
	for l.pos+n < len(l.line) && match(l.line[l.pos+n]) {
		n++
	}
	if n > 0 {
		l.take(n, kind)
	}
}

func (l *lexer) hasPrefix(s string) bool {
	pattern := []rune(s)
	rest := l.line[l.pos:]
	return len(rest) >= len(pattern) && slices.Equal(rest[:len(pattern)], pattern)
}

func index(runes []rune, s string) int {
	pattern := []rune(s)
	for i := 0; i+len(pattern) <= len(runes); i++ {
		if slices.Equal(runes[i:i+len(pattern)], pattern) {
			return i
		}
	}
	return -1
}

func isNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_:-.", r)
}
