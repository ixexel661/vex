package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/editor"
)

type foundPairMsg editor.FoundPair

// findPair looks for the other tag of a pair in the background, if it is far
// from the cursor, one job at a time.
func (m *Model) findPair() tea.Cmd {
	if m.findingPair {
		return nil
	}
	job := m.editor.FindPair()
	if job == nil {
		return nil
	}
	m.findingPair = true
	return func() tea.Msg {
		return foundPairMsg(job())
	}
}

func (m *Model) foundPair(msg foundPairMsg) {
	m.findingPair = false
	m.editor.ApplyFoundPair(editor.FoundPair(msg))
}
