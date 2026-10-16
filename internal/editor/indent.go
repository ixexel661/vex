package editor

import (
	"strings"

	"github.com/ixexel661/vex/internal/buffer"
)

// Indent inserts spaces up to the next indent stop, or a tab if the config
// indents with tabs.
func (e *Editor) Indent() {
	if e.cfg.Format.Indent == "\t" {
		e.InsertText("\t")
		return
	}
	n := e.indentWidth()
	e.InsertText(strings.Repeat(" ", n-e.cursorColumn()%n))
}

// Outdent takes the indentation of the cursor's line back to the previous
// indent stop. A tab goes as a whole. The cursor stays on its text.
func (e *Editor) Outdent() {
	row := e.cursor.Row
	end := leadingWhitespace([]rune(e.buffer().Line(row)))
	width := e.column(buffer.Position{Row: row, Col: end})
	if width == 0 {
		return
	}
	stop := e.stopBefore(width)
	start := end
	for e.column(buffer.Position{Row: row, Col: start}) > stop {
		start--
	}
	e.change(buffer.Range{Start: buffer.Position{Row: row, Col: start}, End: buffer.Position{Row: row, Col: end}}, "")
	switch {
	case e.cursor.Col >= end:
		e.cursor.Col -= end - start
	case e.cursor.Col > start:
		e.cursor.Col = start
	}
	e.setCursor(e.cursor)
}

// backspaceStart returns where backspace starts to delete: at the rune
// before the cursor, or in the indentation, at the previous indent stop as
// far as there are spaces.
func (e *Editor) backspaceStart() buffer.Position {
	row, col := e.cursor.Row, e.cursor.Col
	line := []rune(e.buffer().Line(row))
	if col == 0 || leadingWhitespace(line) < col || line[col-1] != ' ' {
		return e.buffer().Prev(e.cursor)
	}
	stop := e.stopBefore(e.cursorColumn())
	from := col
	for from > 0 && line[from-1] == ' ' && e.column(buffer.Position{Row: row, Col: from}) > stop {
		from--
	}
	return buffer.Position{Row: row, Col: from}
}

// stopBefore returns the last indent stop before the screen column col.
func (e *Editor) stopBefore(col int) int {
	n := e.indentWidth()
	return (col - 1) / n * n
}

// indentWidth is the number of screen columns of one indent level.
func (e *Editor) indentWidth() int {
	if e.cfg.Format.Indent == "\t" {
		return e.cfg.TabWidth
	}
	return len(e.cfg.Format.Indent)
}

func leadingWhitespace(line []rune) int {
	n := 0
	for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
		n++
	}
	return n
}
