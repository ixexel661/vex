package xml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/ixexel661/vex/internal/config"
)

// Formatter re-indents a document. Only whitespace between tags changes:
// elements that contain text are kept exactly as written.
type Formatter struct{}

func (Formatter) Format(src string, opts config.Format) (string, error) {
	nodes, err := parse(src)
	if err != nil {
		return "", err
	}
	p := printer{opts: opts}
	for _, n := range nodes {
		p.print(n, 0)
	}
	return p.String(), nil
}

// node is a part of the document together with its original source text.
type node struct {
	source      string
	startTag    string
	endTag      string
	start       int
	children    []*node
	isElement   bool
	hasText     bool
	blankBefore bool // a blank line separates it from the node before
}

func parse(src string) ([]*node, error) {
	d := newDecoder(src)
	document := &node{}
	open := []*node{document}
	offset := 0
	blank := false
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			return document.children, nil
		}
		if err != nil {
			return nil, toSyntaxError(err)
		}
		// The decoder only hands out parsed tokens, so the original text of
		// each token is cut out of the source between two input offsets.
		end := int(d.InputOffset())
		raw := src[offset:end]
		parent := open[len(open)-1]
		switch tok.(type) {
		case stdxml.StartElement:
			n := &node{startTag: raw, start: offset, isElement: true}
			parent.add(n, blank)
			open = append(open, n)
		case stdxml.EndElement:
			parent.endTag = raw
			parent.source = src[parent.start:end]
			open = open[:len(open)-1]
		case stdxml.CharData:
			if strings.TrimSpace(raw) != "" {
				parent.hasText = true
				parent.add(&node{source: raw}, blank)
			}
		default:
			parent.add(&node{source: raw}, blank)
		}
		blank = strings.TrimSpace(raw) == "" && strings.Count(raw, "\n") > 1
		offset = end
	}
}

// add appends child. blank says if a blank line comes before it in the
// source, which only counts between children.
func (n *node) add(child *node, blank bool) {
	child.blankBefore = blank && len(n.children) > 0
	n.children = append(n.children, child)
}

type printer struct {
	strings.Builder
	opts config.Format
}

func (p *printer) print(n *node, depth int) {
	if n.blankBefore && p.opts.KeepBlankLines {
		p.WriteString("\n")
	}
	switch {
	case !n.isElement:
		p.line(depth, strings.TrimSpace(n.source))
	case len(n.children) == 0 && p.opts.SelfCloseEmpty:
		p.line(depth, p.tag(selfClosing(n.startTag), depth))
	case n.hasText || len(n.children) == 0:
		// Only the start tag may change, the rest stays as written.
		p.line(depth, p.tag(n.startTag, depth)+n.source[len(n.startTag):])
	default:
		p.line(depth, p.tag(n.startTag, depth))
		for _, child := range n.children {
			p.print(child, depth+1)
		}
		p.line(depth, n.endTag)
	}
}

// tag lays out a start tag. Without a limit it stays as written. Otherwise
// it goes on one line, or if that is longer than the limit, each attribute
// goes on a line of its own.
func (p *printer) tag(raw string, depth int) string {
	limit := p.opts.WrapTagsLongerThan
	if limit == 0 {
		return raw
	}
	name, attrs, end := splitTag(raw)
	head := "<" + name
	line := strings.Join(append([]string{head}, attrs...), " ") + end
	if len(attrs) == 0 || utf8.RuneCountInString(line) <= limit {
		return line
	}
	indent := "\n" + strings.Repeat(string(p.opts.Indent), depth+1)
	return head + indent + strings.Join(attrs, indent) + end
}

func (p *printer) line(depth int, text string) {
	p.WriteString(strings.Repeat(string(p.opts.Indent), depth))
	p.WriteString(text)
	p.WriteString("\n")
}
