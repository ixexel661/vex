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

// Setup is what the model is built from. main.go fills it in.
type Setup struct {
	Commands *command.Registry
	Keys     *keymap.Keymap
	// Matchers are the search modes. Tab switches between them in this
	// order, the first is the default.
	Matchers   []search.Matcher
	Config     config.Config
	ConfigPath string // where the settings popup saves the config
	Theme      theme.Theme
	Themes     *theme.Store
	// Warning is shown in the message line at the start, e.g. a broken
	// config file.
	Warning error
	// NewFile starts in insert mode at the end of the file type's template.
	NewFile bool
	// Version is the version of vex, shown in the settings.
	Version string
}

type Model struct {
	editor     *editor.Editor
	commands   *command.Registry
	keys       *keymap.Keymap
	cfg        config.Config
	configPath string
	theme      theme.Theme
	themes     *theme.Store
	version    string

	mode          mode.Mode
	width, height int
	message       message
	dialog        *dialog
	prompt        *prompt
	search        *search.Session
	searching     bool
	replaceQuery  bool
	settings      *settingsPopup
	validation    validation
	quitting      bool
}

type message struct {
	text    string
	isError bool
}

func New(doc *document.Document, s Setup) *Model {
	m := &Model{
		editor:     editor.New(doc, s.Config, s.Theme),
		commands:   s.Commands,
		keys:       s.Keys,
		cfg:        s.Config,
		configPath: s.ConfigPath,
		theme:      s.Theme,
		themes:     s.Themes,
		version:    s.Version,
		search:     search.NewSession(doc.Buffer, s.Matchers),
		validation: validation{version: -1},
	}
	if s.NewFile {
		m.SetMode(mode.Insert)
		m.editor.MoveBottom()
	}
	m.reportError(s.Warning)
	m.validate()
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}
