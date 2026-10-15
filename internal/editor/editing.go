package editor

import "strings"

func (e *Editor) InsertText(text string) {
	e.edit(func() {
		e.cursor = e.buffer().Insert(e.cursor, text)
	})
}

// Newline breaks the line at the cursor and keeps its indentation.
func (e *Editor) Newline() {
	before := string([]rune(e.buffer().Line(e.cursor.Row))[:e.cursor.Col])
	indent := before[:len(before)-len(strings.TrimLeft(before, " \t"))]
	e.InsertText("\n" + indent)
}

func (e *Editor) DeleteBackward() {
	e.edit(func() {
		prev := e.buffer().Prev(e.cursor)
		e.buffer().Delete(prev, e.cursor)
		e.cursor = prev
	})
}

func (e *Editor) DeleteForward() {
	e.edit(func() {
		e.buffer().Delete(e.cursor, e.buffer().Next(e.cursor))
	})
}

// edit applies a change at the cursor. Highlighting is refreshed from the
// first row the change touched, which is the cursor row before or after it.
func (e *Editor) edit(change func()) {
	row := e.cursor.Row
	change()
	e.highlights.Invalidate(min(row, e.cursor.Row))
	e.setCursor(e.cursor)
}
