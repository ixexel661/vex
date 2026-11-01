package app

import (
	"strings"

	"github.com/ixexel661/vex/internal/printable"
)

func (m *Model) messageLine() string {
	switch {
	case m.dialog != nil:
		return m.question()
	case m.prompt != nil:
		return m.promptLine()
	case m.searching:
		return m.searchBar()
	case m.problemList != nil:
		return m.problemBar()
	case m.message.isError:
		return m.theme.Error.Render(printable.String(m.message.text))
	case m.message.text != "":
		return m.theme.Message.Render(printable.String(m.message.text))
	case m.schemaPopup != nil:
		return m.theme.Hint.Render(schemaHints)
	case m.settings != nil:
		return m.theme.Hint.Render(settingsHints)
	case m.editor.Jumping():
		return m.theme.Hint.Render(jumpHint)
	}
	if p, ok := m.editor.ProblemAt(m.editor.Cursor().Row); ok {
		return m.theme.Error.Render(printable.String(p.Message))
	}
	return m.theme.Hint.Render(m.hints())
}

func (m *Model) question() string {
	var options []string
	for _, c := range m.dialog.choices {
		options = append(options, c.Key+" "+c.Label)
	}
	options = append(options, "esc cancel")
	return m.theme.Question.Render(printable.String(m.dialog.question)) + "  " + m.theme.Hint.Render(strings.Join(options, " · "))
}

func (m *Model) hints() string {
	var hints []string
	shown := map[string]bool{}
	for _, b := range m.keys.Bindings(m.mode) {
		c, ok := m.commands.Lookup(b.Command)
		if !ok || c.Title == "" || shown[c.Name] {
			continue
		}
		shown[c.Name] = true
		hints = append(hints, b.Key+" "+c.Title)
	}
	return strings.Join(hints, " · ")
}
