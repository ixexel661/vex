// Package sample makes large documents for tests and benchmarks.
package sample

import (
	"fmt"
	"strings"
)

// XML returns a well-formed document of about size bytes: a library of
// books with attributes, text, comments, CDATA and some non-ASCII text.
func XML(size int) string {
	var b strings.Builder
	b.Grow(size + 1024)
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<library>\n")
	for i := 1; b.Len() < size; i++ {
		fmt.Fprintf(&b, `  <!-- book %d -->
  <book id="%d" lang="de">
    <title>Über das Buch %d</title>
    <author>Autor %d</author>
    <price currency="EUR">%d.50</price>
    <description><![CDATA[Some <b>text</b> & more]]></description>
    <tags>
      <tag>xml</tag>
      <tag>go</tag>
    </tags>
  </book>
`, i, i, i, i%97, i%50)
	}
	b.WriteString("</library>\n")
	return b.String()
}
