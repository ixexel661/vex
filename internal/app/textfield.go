package app

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/printable"
)

// textField is a line of text with a cursor, e.g. in a popup.
type textField struct {
	text   []rune
	cursor int
}

func newTextField(text string) *textField {
	f := &textField{}
	f.set(text)
	return f
}

func (f *textField) String() string {
	return string(f.text)
}

// set replaces the text and puts the cursor at its end.
func (f *textField) set(text string) {
	f.text = []rune(text)
	f.cursor = len(f.text)
}

func (f *textField) insert(text string) {
	runes := []rune(text)
	f.text = slices.Insert(f.text, f.cursor, runes...)
	f.cursor += len(runes)
}

// key moves the cursor, deletes or types.
func (f *textField) key(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "left":
		f.cursor = max(f.cursor-1, 0)
	case "right":
		f.cursor = min(f.cursor+1, len(f.text))
	case "home":
		f.cursor = 0
	case "end":
		f.cursor = len(f.text)
	case "backspace":
		if f.cursor > 0 {
			f.text = slices.Delete(f.text, f.cursor-1, f.cursor)
			f.cursor--
		}
	case "delete":
		if f.cursor < len(f.text) {
			f.text = slices.Delete(f.text, f.cursor, f.cursor+1)
		}
	default:
		if msg.Text != "" {
			f.insert(msg.Text)
		}
	}
}

// view shows as much of the text as fits in width runes, scrolled so the
// cursor shows, and returns where the cursor is in it.
func (f *textField) view(width int) (string, int) {
	width = max(width, 1)
	first := max(f.cursor-width+1, 0)
	last := min(first+width, len(f.text))
	return printable.String(string(f.text[first:last])), f.cursor - first
}
