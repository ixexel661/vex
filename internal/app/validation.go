package app

import (
	"errors"
	"fmt"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/mode"
)

type validation struct {
	version int // buffer version the result belongs to
	schemas int // schemas generation the result belongs to
	err     error
	// schema reports whether the text was checked against schemas,
	// schemaErr why they could not be loaded.
	schema    bool
	schemaErr error
	running   bool
}

// validatedMsg is the result of a check that ran in the background.
type validatedMsg struct {
	version, schemas int
	validity         document.Validity
}

// validate starts a check of the text in the background if it or its
// schemas changed since the last one. Only one check runs at a time.
func (m *Model) validate() tea.Cmd {
	doc := m.editor.Document()
	if m.validation.running || m.mode != mode.Normal ||
		doc.Buffer.Version() == m.validation.version && m.schemas == m.validation.schemas {
		return nil
	}
	m.validation.running = true
	version, schemas, check := doc.Buffer.Version(), m.schemas, doc.Validation()
	return func() tea.Msg {
		return validatedMsg{version: version, schemas: schemas, validity: check()}
	}
}

// validated keeps the result of a check, unless the text or the schemas
// changed meanwhile. Then the next call of validate checks again.
func (m *Model) validated(msg validatedMsg) {
	m.validation.running = false
	doc := m.editor.Document()
	if msg.version != doc.Buffer.Version() || msg.schemas != m.schemas {
		return
	}
	v := msg.validity
	m.validation = validation{version: msg.version, schemas: msg.schemas, err: v.Err, schema: v.Schema, schemaErr: v.SchemaErr}
	problems := v.Problems
	if v.Schema && v.Err != nil {
		problems = m.syntaxProblem(v.Err)
	}
	m.editor.SetProblems(problems)
	if m.message.text != "" {
		return
	}
	switch {
	case v.Err != nil:
		m.message = message{text: fmt.Sprintf("Invalid %s: %v", doc.Type.Name, v.Err), isError: true}
	case v.SchemaErr != nil:
		m.message = message{text: fmt.Sprintf("Cannot load the schema: %v", v.SchemaErr), isError: true}
	}
}

// syntaxProblem marks the line of a syntax error.
func (m *Model) syntaxProblem(err error) []filetype.Problem {
	var syntaxErr *filetype.SyntaxError
	buf := m.editor.Document().Buffer
	if !errors.As(err, &syntaxErr) || syntaxErr.Line < 1 || syntaxErr.Line > buf.LineCount() {
		return nil
	}
	row := syntaxErr.Line - 1
	line := []rune(buf.Line(row))
	start, end := 0, len(line)
	for start < end && unicode.IsSpace(line[start]) {
		start++
	}
	for end > start && unicode.IsSpace(line[end-1]) {
		end--
	}
	return []filetype.Problem{{
		Range:   buffer.Range{Start: buffer.Position{Row: row, Col: start}, End: buffer.Position{Row: row, Col: end}},
		Message: syntaxErr.Message,
	}}
}
