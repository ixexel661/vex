package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/mode"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		m.handleKey(msg)
	case tea.PasteMsg:
		m.handlePaste(msg.Content)
	}
	m.layout()
	m.validate()
	if m.quitting {
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) layout() {
	m.editor.SetSize(m.width, m.editorHeight())
}

// editorHeight leaves room for the match list, the status bar and the
// message line.
func (m *Model) editorHeight() int {
	return m.height - 2 - m.listHeight()
}

func (m *Model) handleKey(msg tea.KeyPressMsg) {
	m.message = message{}
	key := msg.String()
	if m.dialog != nil {
		m.answer(key)
		return
	}
	if m.searching {
		m.searchKey(msg)
		return
	}
	if m.settings != nil {
		m.settingsKey(key)
		return
	}
	if name, ok := m.keys.Lookup(m.mode, key); ok {
		m.run(name)
		return
	}
	if m.mode == mode.Insert && msg.Text != "" {
		m.editor.InsertText(msg.Text)
	}
}

func (m *Model) handlePaste(text string) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	switch {
	case m.searching:
		line, _, _ := strings.Cut(text, "\n")
		m.typeInSearch(line)
	case m.mode == mode.Insert:
		m.editor.InsertText(text)
	default:
		m.Notify("Press i to paste text")
	}
}

func (m *Model) run(name string) {
	c, ok := m.commands.Lookup(name)
	if !ok {
		m.reportError(fmt.Errorf("unknown command %q", name))
		return
	}
	m.reportError(c.Run(m))
}

func (m *Model) reportError(err error) {
	if err != nil {
		// Joined errors come one per line, the message line has only one.
		m.message = message{text: strings.ReplaceAll(err.Error(), "\n", "; "), isError: true}
	}
}
