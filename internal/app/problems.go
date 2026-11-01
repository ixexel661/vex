package app

import (
	"fmt"

	"github.com/ixexel661/vex/internal/buffer"
)

// problemList is the open list of the problems in the text.
type problemList struct {
	selected int
	origin   buffer.Position // where the cursor goes back to on cancel
}

// ShowProblems lists the problems against the schemas, or opens the schema
// popup if the document has none yet.
func (m *Model) ShowProblems() {
	doc := m.editor.Document()
	switch {
	case doc.Type.Schemas == nil:
		m.Notify(fmt.Sprintf("%s files have no schemas", doc.Type.Name))
	case doc.Schemas() == nil:
		m.OpenSchemas()
	case doc.SchemasChanged():
		m.schemas++
		m.Notify("The schema changed, checking again")
	case len(m.editor.Problems()) > 0:
		m.openProblems()
	case m.validation.running:
		m.Notify("Still checking")
	case m.validation.schemaErr != nil:
		m.reportError(fmt.Errorf("cannot load the schema: %w", m.validation.schemaErr))
	default:
		m.Notify("No errors against the schema")
	}
}

// openProblems selects the first problem from the cursor on.
func (m *Model) openProblems() {
	cursor := m.editor.Cursor()
	selected := 0
	for i, p := range m.editor.Problems() {
		if !p.Range.Start.Before(cursor) {
			selected = i
			break
		}
	}
	m.problemList = &problemList{selected: selected, origin: cursor}
	m.showProblem()
}

func (m *Model) problemKey(key string) {
	switch key {
	case "up":
		m.moveInProblems(-1)
	case "down":
		m.moveInProblems(1)
	case "enter":
		m.problemList = nil
	case "esc", "ctrl+c":
		m.editor.MoveTo(m.problemList.origin)
		m.problemList = nil
	}
}

func (m *Model) moveInProblems(delta int) {
	if n := len(m.editor.Problems()); n > 0 {
		m.problemList.selected = ((m.problemList.selected+delta)%n + n) % n
		m.showProblem()
	}
}

// showProblem puts the cursor at the start of the selected problem.
func (m *Model) showProblem() {
	if problems := m.editor.Problems(); m.problemList.selected < len(problems) {
		m.editor.MoveTo(problems[m.problemList.selected].Range.Start)
	}
}
