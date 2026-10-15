package app

import (
	"fmt"

	"github.com/ixexel661/vex/internal/mode"
)

type validation struct {
	version int // buffer version the result belongs to
	err     error
}

// validate checks the document after it changed. It waits for normal mode,
// because while typing the document is often invalid for a moment.
func (m *Model) validate() {
	doc := m.editor.Document()
	if m.mode != mode.Normal || doc.Buffer.Version() == m.validation.version {
		return
	}
	m.validation = validation{version: doc.Buffer.Version(), err: doc.Validate()}
	if m.validation.err != nil && m.message.text == "" {
		m.message = message{text: fmt.Sprintf("Invalid %s: %v", doc.Type.Name, m.validation.err), isError: true}
	}
}
