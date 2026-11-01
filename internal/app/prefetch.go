package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/editor"
)

type prefetchedMsg editor.Prefetched

// prefetch highlights the rest of the document in the background, one job
// at a time.
func (m *Model) prefetch() tea.Cmd {
	if m.prefetching {
		return nil
	}
	job := m.editor.Prefetch()
	if job == nil {
		return nil
	}
	m.prefetching = true
	return func() tea.Msg {
		return prefetchedMsg(job())
	}
}

func (m *Model) prefetched(msg prefetchedMsg) {
	m.prefetching = false
	m.editor.ApplyPrefetched(editor.Prefetched(msg))
}
