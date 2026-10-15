package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/command"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/editor"
	"github.com/ixexel661/vex/internal/keymap"
	"github.com/ixexel661/vex/internal/mode"
	"github.com/ixexel661/vex/internal/search"
	"github.com/ixexel661/vex/internal/theme"
)

// Model is the Bubble Tea model of vex. It routes keys to commands and lays
// out the editor, the list of search matches, the status bar and the message
// line.
type Model struct {
	editor   *editor.Editor
	commands *command.Registry
	keys     *keymap.Keymap
	cfg      config.Config
	theme    theme.Theme

	mode          mode.Mode
	width, height int
	message       message
	dialog        *dialog
	search        *search.Session
	searching     bool
	replaceQuery  bool // the query is still selected from the last search
	validation    validation
	quitting      bool
}

type message struct {
	text    string
	isError bool
}

func New(doc *document.Document, commands *command.Registry, keys *keymap.Keymap, matchers []search.Matcher, cfg config.Config, th theme.Theme) *Model {
	m := &Model{
		editor:     editor.New(doc, cfg, th),
		commands:   commands,
		keys:       keys,
		cfg:        cfg,
		theme:      th,
		search:     search.NewSession(doc.Buffer, matchers),
		validation: validation{version: -1},
	}
	m.validate()
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}
