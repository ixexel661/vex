package editor

import "github.com/ixexel661/vex/internal/buffer"

func (e *Editor) MoveLeft() {
	e.setCursor(e.buffer().Prev(e.cursor))
}

func (e *Editor) MoveRight() {
	e.setCursor(e.buffer().Next(e.cursor))
}

func (e *Editor) MoveUp() {
	e.moveRows(-1)
}

func (e *Editor) MoveDown() {
	e.moveRows(1)
}

func (e *Editor) PageUp() {
	e.moveRows(-e.height)
}

func (e *Editor) PageDown() {
	e.moveRows(e.height)
}

func (e *Editor) MoveLineStart() {
	e.setCursor(buffer.Position{Row: e.cursor.Row})
}

func (e *Editor) MoveLineEnd() {
	e.setCursor(buffer.Position{Row: e.cursor.Row, Col: e.buffer().LineLen(e.cursor.Row)})
}

func (e *Editor) MoveTop() {
	e.setCursor(buffer.Position{})
}

func (e *Editor) MoveBottom() {
	last := e.buffer().LineCount() - 1
	e.setCursor(buffer.Position{Row: last, Col: e.buffer().LineLen(last)})
}

// MoveTo puts the cursor at p, or as close to it as the text allows.
func (e *Editor) MoveTo(p buffer.Position) {
	e.setCursor(e.buffer().Clamp(p))
}

func (e *Editor) moveRows(delta int) {
	e.cursor = e.buffer().Clamp(buffer.Position{Row: e.cursor.Row + delta, Col: e.goalCol})
	e.scrollToCursor()
}

func (e *Editor) setCursor(p buffer.Position) {
	e.cursor = p
	e.goalCol = p.Col
	e.scrollToCursor()
}
