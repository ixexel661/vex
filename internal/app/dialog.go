package app

import "github.com/ixexel661/vex/internal/command"

type dialog struct {
	question string
	choices  []command.Choice
}

func (m *Model) Ask(question string, choices ...command.Choice) {
	m.dialog = &dialog{question: question, choices: choices}
}

// answer handles a key while a question is open. Other keys are ignored, so
// a stray key press cannot pick an answer by accident.
func (m *Model) answer(key string) {
	if key == "esc" || key == "ctrl+c" {
		m.dialog = nil
		return
	}
	for _, c := range m.dialog.choices {
		if c.Key == key {
			m.dialog = nil
			m.reportError(c.Run(m))
			return
		}
	}
}
