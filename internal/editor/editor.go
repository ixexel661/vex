package editor

import (
	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/theme"
)

// Editor shows a document in a scrollable window and edits it at the cursor.
type Editor struct {
	doc        *document.Document
	cfg        config.Config
	theme      theme.Theme
	highlights *highlight.Cache
	marks      map[int][]highlight.Span // by row

	cursor  buffer.Position
	goalCol int // the column vertical movement returns to on long enough lines

	top, left     int // first visible row and screen column of the text
	width, height int
}

func New(doc *document.Document, cfg config.Config, th theme.Theme) *Editor {
	return &Editor{
		doc:        doc,
		cfg:        cfg,
		theme:      th,
		highlights: highlight.NewCache(doc.Type.Highlighter),
		width:      1,
		height:     1,
	}
}

func (e *Editor) Document() *document.Document {
	return e.doc
}

func (e *Editor) Cursor() buffer.Position {
	return e.cursor
}

func (e *Editor) SetSize(width, height int) {
	e.width, e.height = max(width, 1), max(height, 1)
	e.scrollToCursor()
}

// SetText replaces the whole text, e.g. after formatting.
func (e *Editor) SetText(text string) {
	e.buffer().SetText(text)
	e.highlights.Invalidate(0)
	e.setCursor(e.buffer().Clamp(e.cursor))
}

func (e *Editor) buffer() *buffer.Buffer {
	return e.doc.Buffer
}

func (e *Editor) scrollToCursor() {
	e.top = clamp(e.top, e.cursor.Row-e.height+1, e.cursor.Row)
	x := e.cursorColumn()
	e.left = clamp(e.left, x-e.textWidth()+1, x)
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
