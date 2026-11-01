package editor

import (
	"strings"
	"unicode/utf8"

	"github.com/ixexel661/vex/internal/buffer"
)

func (e *Editor) InsertText(text string) {
	e.replace(e.cursor, e.cursor, text)
}

// Newline breaks the line at the cursor and keeps its indentation.
func (e *Editor) Newline() {
	before := string([]rune(e.buffer().Line(e.cursor.Row))[:e.cursor.Col])
	indent := before[:len(before)-len(strings.TrimLeft(before, " \t"))]
	e.InsertText("\n" + indent)
}

// DeleteBackward deletes the rune before the cursor. In the indentation, it
// deletes spaces back to the previous indent stop.
func (e *Editor) DeleteBackward() {
	e.replace(e.backspaceStart(), e.cursor, "")
}

func (e *Editor) DeleteForward() {
	e.replace(e.cursor, e.buffer().Next(e.cursor), "")
}

// replace puts text in place of the text from "from" up to "to" at the
// cursor and leaves the cursor after it. With linked editing, a change to a
// tag name also happens in the other name of the pair, which then stays
// known. The later of the two changes comes first, so it doesn't move the
// other one.
func (e *Editor) replace(from, to buffer.Position, text string) {
	twin, linked := e.twin(from, to, text)
	twinAfter := linked && from.Before(twin.Start)
	if twinAfter {
		e.change(twin, text)
	}
	e.cursor = e.change(buffer.Range{Start: from, End: to}, text)
	if linked && !twinAfter {
		e.change(twin, text)
		if twin.Start.Row == e.cursor.Row {
			e.cursor.Col += utf8.RuneCountInString(text) - (twin.End.Col - twin.Start.Col)
		}
	}
	if linked {
		e.knownPair = e.knownPair.renamed(from, to, text, e.buffer().Version())
	}
	e.setCursor(e.cursor)
}

// change puts text in place of r and returns where the text ends.
func (e *Editor) change(r buffer.Range, text string) buffer.Position {
	e.buffer().Delete(r.Start, r.End)
	end := r.Start
	if text != "" {
		end = e.buffer().Insert(r.Start, text)
	}
	e.highlights.Edit(e.buffer(), r.Start.Row, r.End.Row-r.Start.Row, end.Row-r.Start.Row)
	e.moveProblems(r.Start, r.End, end)
	return end
}
