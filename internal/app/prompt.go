package app

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/command"
)

const promptHints = "enter confirm · esc cancel"

// prompt is a line of text typed into the message line, e.g. a file name.
type prompt struct {
	label string
	text  string
	run   func(ctx command.Context, text string) error
}

func (m *Model) Prompt(label string, run func(ctx command.Context, text string) error) {
	m.prompt = &prompt{label: label, run: run}
}

func (m *Model) promptKey(msg tea.KeyPressMsg) {
	p := m.prompt
	switch msg.String() {
	case "enter":
		if p.text != "" {
			m.prompt = nil
			m.reportError(p.run(m, p.text))
		}
	case "esc", "ctrl+c":
		m.prompt = nil
	case "backspace":
		_, size := utf8.DecodeLastRuneInString(p.text)
		p.text = p.text[:len(p.text)-size]
	default:
		m.typeInPrompt(msg.Text)
	}
}

func (m *Model) typeInPrompt(text string) {
	m.prompt.text += text
}

// promptInput is the part of the prompt line up to the end of the text.
func (m *Model) promptInput() string {
	return m.theme.Question.Render(m.prompt.label) + " " + printable(m.prompt.text)
}

func (m *Model) promptLine() string {
	return m.promptInput() + "  " + m.theme.Hint.Render(promptHints)
}
