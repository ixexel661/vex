package xml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"strings"
)

// Formatter re-indents a document. Only whitespace between tags changes:
// elements that contain text are kept exactly as written.
type Formatter struct{}

func (Formatter) Format(src, indent string) (string, error) {
	nodes, err := parse(src)
	if err != nil {
		return "", err
	}
	p := printer{indent: indent}
	for _, n := range nodes {
		p.print(n, 0)
	}
	return p.String(), nil
}

// node is a part of the document together with its original source text.
type node struct {
	source    string
	startTag  string
	endTag    string
	start     int
	children  []*node
	isElement bool
	hasText   bool
}

func parse(src string) ([]*node, error) {
	d := newDecoder(src)
	document := &node{}
	open := []*node{document}
	offset := 0
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
			parent.children = append(parent.children, n)
			open = append(open, n)
		case stdxml.EndElement:
			parent.endTag = raw
			parent.source = src[parent.start:end]
			open = open[:len(open)-1]
		case stdxml.CharData:
			if strings.TrimSpace(raw) != "" {
				parent.hasText = true
				parent.children = append(parent.children, &node{source: raw})
			}
		default:
			parent.children = append(parent.children, &node{source: raw})
		}
		offset = end
	}
}

type printer struct {
	strings.Builder
	indent string
}

func (p *printer) print(n *node, depth int) {
	if !n.isElement || n.hasText || len(n.children) == 0 {
		p.line(depth, strings.TrimSpace(n.source))
		return
	}
	p.line(depth, n.startTag)
	for _, child := range n.children {
		p.print(child, depth+1)
	}
	p.line(depth, n.endTag)
}

func (p *printer) line(depth int, text string) {
	p.WriteString(strings.Repeat(p.indent, depth))
	p.WriteString(text)
	p.WriteString("\n")
}
