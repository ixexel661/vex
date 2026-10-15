package app

import (
	"github.com/ixexel661/vex/internal/command"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/editor"
	"github.com/ixexel661/vex/internal/mode"
)

var _ command.Context = (*Model)(nil)

func (m *Model) Editor() *editor.Editor {
	return m.editor
}

func (m *Model) Config() config.Config {
	return m.cfg
}

func (m *Model) SetMode(md mode.Mode) {
	m.mode = md
}

func (m *Model) Notify(text string) {
	m.message = message{text: text}
}

func (m *Model) Quit() {
	m.quitting = true
}
