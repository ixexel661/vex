package editor

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/highlight"
)

// cell is a piece of a line as it appears on screen.
type cell struct {
	text  string
	width int
	index int // index of the rune in the line the cell shows
}

func (e *Editor) View() string {
	rows := make([]string, e.height)
	for y := range rows {
		row := e.top + y
		if row >= e.buffer().LineCount() {
			rows[y] = e.theme.Filler.Render("~")
			continue
		}
		rows[y] = e.gutter(row) + e.renderLine(row)
	}
	return strings.Join(rows, "\n")
}

// CursorPosition returns where the cursor is on screen, relative to the
// top left corner of the editor.
func (e *Editor) CursorPosition() (x, y int) {
	return e.gutterWidth() + e.cursorColumn() - e.left, e.cursor.Row - e.top
}

func (e *Editor) gutterWidth() int {
	return max(len(strconv.Itoa(e.buffer().LineCount())), 3) + 1
}

func (e *Editor) textWidth() int {
	return max(e.width-e.gutterWidth(), 1)
}

func (e *Editor) gutter(row int) string {
	style := e.theme.LineNumber
	if row == e.cursor.Row {
		style = e.theme.CurrentLineNumber
	}
	return style.Render(fmt.Sprintf("%*d ", e.gutterWidth()-1, row+1))
}

// renderLine highlights a row. The marks come after the syntax spans, so
// they win where both cover a rune.
func (e *Editor) renderLine(row int) string {
	line := []rune(e.buffer().Line(row))
	kinds := kindsOf(len(line), slices.Concat(e.highlights.Spans(e.buffer(), row), e.marks[row]))
	return e.paint(visible(e.cells(line), e.left, e.textWidth()), kinds)
}

func (e *Editor) cursorColumn() int {
	line := []rune(e.buffer().Line(e.cursor.Row))
	col := 0
	for _, c := range e.cells(line[:e.cursor.Col]) {
		col += c.width
	}
	return col
}

// cells expands tabs to the next tab stop and replaces other control
// characters, which would garble the screen.
func (e *Editor) cells(line []rune) []cell {
	var cells []cell
	col := 0
	for i, r := range line {
		switch {
		case r == '\t':
			for range e.cfg.TabWidth - col%e.cfg.TabWidth {
				cells = append(cells, cell{text: " ", width: 1, index: i})
				col++
			}
		case unicode.IsControl(r):
			cells = append(cells, cell{text: "?", width: 1, index: i})
			col++
		default:
			c := cell{text: string(r), width: lipgloss.Width(string(r)), index: i}
			cells = append(cells, c)
			col += c.width
		}
	}
	return cells
}

// visible returns the cells within the screen columns [left, left+width).
// A wide character cut by the left edge shows as spaces.
func visible(cells []cell, left, width int) []cell {
	var out []cell
	col := 0
	for _, c := range cells {
		start := col
		col += c.width
		switch {
		case col <= left:
			continue
		case col > left+width:
			return out
		case start < left:
			out = append(out, cell{text: strings.Repeat(" ", col-left), width: col - left, index: c.index})
		default:
			out = append(out, c)
		}
	}
	return out
}

func kindsOf(length int, spans []highlight.Span) []highlight.Kind {
	kinds := make([]highlight.Kind, length)
	for _, s := range spans {
		for i := s.Start; i < s.End && i < length; i++ {
			kinds[i] = s.Kind
		}
	}
	return kinds
}

// paint styles each run of cells that share the same highlight kind.
func (e *Editor) paint(cells []cell, kinds []highlight.Kind) string {
	var out strings.Builder
	for len(cells) > 0 {
		kind := kinds[cells[0].index]
		n := 1
		for n < len(cells) && kinds[cells[n].index] == kind {
			n++
		}
		var text strings.Builder
		for _, c := range cells[:n] {
			text.WriteString(c.text)
		}
		out.WriteString(e.theme.Token(kind).Render(text.String()))
		cells = cells[n:]
	}
	return out.String()
}
