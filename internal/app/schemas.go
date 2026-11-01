package app

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
)

// namedRows is how many rows from the start OpenSchemas reads for the
// schemas a document names itself.
const namedRows = 1000

// schemaPopup edits the schemas of the document: pairs of a namespace and
// a schema file.
type schemaPopup struct {
	fields  []*textField // the namespace and the file of each schema
	focused int
	err     error // why the schemas cannot be used
}

// OpenSchemas opens the schema popup with the schemas of the document, or
// the ones the document names itself.
func (m *Model) OpenSchemas() {
	doc := m.editor.Document()
	if doc.Type.Schemas == nil {
		m.Notify(fmt.Sprintf("%s files have no schemas", doc.Type.Name))
		return
	}
	schemas := doc.Schemas()
	if schemas == nil {
		schemas = doc.Type.Schemas.Named(head(doc.Buffer, namedRows))
	}
	if len(schemas) == 0 {
		schemas = []config.Schema{{}}
	}
	p := &schemaPopup{}
	for _, s := range schemas {
		p.fields = append(p.fields, newTextField(s.Namespace), newTextField(s.Path))
	}
	for p.focused < len(p.fields)-1 && p.fields[p.focused].String() != "" {
		p.focused++
	}
	m.schemaPopup = p
}

// head returns the first rows of b.
func head(b *buffer.Buffer, rows int) string {
	lines := make([]string, min(rows, b.LineCount()))
	for i := range lines {
		lines[i] = b.Line(i)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) schemaKey(msg tea.KeyPressMsg) {
	p := m.schemaPopup
	p.err = nil
	switch msg.String() {
	case "up":
		p.focused = (p.focused + len(p.fields) - 1) % len(p.fields)
	case "down":
		p.focused = (p.focused + 1) % len(p.fields)
	case "tab":
		if p.focused%2 == 1 {
			f := p.fields[p.focused]
			f.set(completePath(f.String(), m.editor.Document().Resolve))
		}
	case "ctrl+n":
		p.fields = append(p.fields, newTextField(""), newTextField(""))
		p.focused = len(p.fields) - 2
	case "ctrl+d":
		p.removeSchema()
	case "enter":
		m.saveSchemas()
	case "esc", "ctrl+c":
		m.schemaPopup = nil
	default:
		p.fields[p.focused].key(msg)
	}
}

// removeSchema removes the schema of the focused field. The last one is
// emptied.
func (p *schemaPopup) removeSchema() {
	first := p.focused / 2 * 2
	if len(p.fields) == 2 {
		p.fields[0].set("")
		p.fields[1].set("")
		p.focused = 0
		return
	}
	p.fields = append(p.fields[:first], p.fields[first+2:]...)
	p.focused = min(first, len(p.fields)-2)
}

// schemas returns the schemas in the popup without the empty ones.
func (p *schemaPopup) schemas() ([]config.Schema, error) {
	var schemas []config.Schema
	for i := 0; i < len(p.fields); i += 2 {
		s := config.Schema{Namespace: strings.TrimSpace(p.fields[i].String()), Path: strings.TrimSpace(p.fields[i+1].String())}
		switch {
		case s == config.Schema{}:
			continue
		case s.Path == "":
			p.focused = i + 1
			return nil, fmt.Errorf("enter the schema file for %s", s.Namespace)
		}
		schemas = append(schemas, s)
	}
	return schemas, nil
}

// saveSchemas sets the schemas of the popup if they load, and stores them
// for the document.
func (m *Model) saveSchemas() {
	p := m.schemaPopup
	schemas, err := p.schemas()
	if err != nil {
		p.err = err
		return
	}
	doc := m.editor.Document()
	old := doc.Schemas()
	doc.SetSchemas(schemas)
	if err := doc.LoadSchemas(); err != nil {
		doc.SetSchemas(old)
		p.err = err
		return
	}
	m.schemaPopup = nil
	m.schemas++
	m.editor.SetProblems(nil)
	if doc.Path == "" || m.schemasPath == "" {
		m.Notify("The schemas apply until vex quits")
		return
	}
	if err := config.SaveSchemas(m.schemasPath, schemasKey(doc), schemas); err != nil {
		m.reportError(fmt.Errorf("cannot save the schemas: %w", err))
		return
	}
	m.Notify("Saved " + filepath.Base(m.schemasPath))
}
