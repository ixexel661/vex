package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/editor"
)

type jumpedMsg editor.Jump

const jumpHint = "… looking for the parent element"

// findJump looks for the tag the cursor waits for in the background, one job
// at a time.
func (m *Model) findJump() tea.Cmd {
	if m.findingJump {
		return nil
	}
	job := m.editor.FindJump()
	if job == nil {
		return nil
	}
	m.findingJump = true
	return func() tea.Msg {
		return jumpedMsg(job())
	}
}

func (m *Model) jumped(msg jumpedMsg) {
	m.findingJump = false
	m.editor.ApplyJump(editor.Jump(msg))
}
