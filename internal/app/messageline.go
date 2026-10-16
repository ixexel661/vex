package app

import "strings"

func (m *Model) messageLine() string {
	switch {
	case m.dialog != nil:
		return m.question()
	case m.prompt != nil:
		return m.promptLine()
	case m.searching:
		return m.searchBar()
	case m.message.isError:
		return m.theme.Error.Render(m.message.text)
	case m.message.text != "":
		return m.theme.Message.Render(m.message.text)
	case m.settings != nil:
		return m.theme.Hint.Render(settingsHints)
	default:
		return m.theme.Hint.Render(m.hints())
	}
}

func (m *Model) question() string {
	var options []string
	for _, c := range m.dialog.choices {
		options = append(options, c.Key+" "+c.Label)
	}
	options = append(options, "esc cancel")
	return m.theme.Question.Render(m.dialog.question) + "  " + m.theme.Hint.Render(strings.Join(options, " · "))
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
