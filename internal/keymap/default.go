package keymap

import (
	"slices"

	"github.com/ixexel661/vex/internal/mode"
)

// Default returns the built-in bindings. Their order decides which keys the
// hints in the message line show.
func Default() *Keymap {
	movement := []Binding{
		{"up", "cursor.up"},
		{"down", "cursor.down"},
		{"left", "cursor.left"},
		{"right", "cursor.right"},
		{"home", "cursor.line-start"},
		{"end", "cursor.line-end"},
		{"pgup", "cursor.page-up"},
		{"pgdown", "cursor.page-down"},
		{"ctrl+home", "cursor.top"},
		{"ctrl+end", "cursor.bottom"},
		{"ctrl+right", "cursor.next-tag"},
		{"ctrl+left", "cursor.prev-tag"},
		{"ctrl+up", "cursor.parent"},
		{"ctrl+down", "cursor.first-child"},
	}
	normal := []Binding{
		{"i", "mode.insert"},
		{"ctrl+s", "document.save"},
		{"f", "document.format"},
		{"ctrl+f", "search.open"},
		{"s", "settings.open"},
		{"q", "app.quit"},
		{"ctrl+c", "app.quit"},
	}
	insert := []Binding{
		{"esc", "mode.normal"},
		{"ctrl+c", "mode.normal"},
		{"ctrl+s", "document.save"},
		{"ctrl+f", "search.open"},
		{"enter", "edit.newline"},
		{"backspace", "edit.delete-backward"},
		{"delete", "edit.delete-forward"},
		{"tab", "edit.indent"},
		{"shift+tab", "edit.outdent"},
		{">", "edit.close-tag"},
	}

	k := New()
	for _, b := range slices.Concat(normal, movement) {
		k.Bind(mode.Normal, b.Key, b.Command)
	}
	for _, b := range slices.Concat(insert, movement) {
		k.Bind(mode.Insert, b.Key, b.Command)
	}
	return k
}
