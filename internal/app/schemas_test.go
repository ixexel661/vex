package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/theme"
)

const notesSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:notes" elementFormDefault="qualified">
  <xs:element name="notes">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="note" type="xs:string" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>
`

// badNotes has two elements the schema does not know, in rows 2 and 4.
const badNotes = `<notes xmlns="urn:notes">
  <note>a</note>
  <todo/>
  <note>b</note>
  <done/>
</notes>`

var (
	shiftV = key("V")
	ctrlN  = tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl}
	ctrlD  = tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl}
)

// writeNext writes a file next to the document at docPath.
func writeNext(t *testing.T, docPath, name, content string) string {
	t.Helper()
	path := filepath.Join(filepath.Dir(docPath), name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// fieldTexts returns the texts in the schema popup.
func fieldTexts(m *Model) []string {
	var texts []string
	for _, f := range m.schemaPopup.fields {
		texts = append(texts, f.String())
	}
	return texts
}

// withNotesSchema returns a model of text whose schema is set to
// notes.xsd in the popup, with the checks done.
func withNotesSchema(t *testing.T, text string) (*Model, string) {
	t.Helper()
	m, path := newModel(t, text)
	writeNext(t, path, "notes.xsd", notesSchema)
	press(m, shiftV)
	press(m, typing("notes.xsd")...)
	settle(m, press(m, enter))
	if m.schemaPopup != nil {
		t.Fatalf("the popup stays open: %v", m.schemaPopup.err)
	}
	return m, path
}

func TestSchemaPopupStartsWithTheNamedSchemas(t *testing.T) {
	tests := []struct {
		name, text string
		want       []string
		focused    int
	}{
		{"schema location", `<notes xmlns="urn:notes" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="urn:notes notes.xsd"/>`, []string{"urn:notes", "notes.xsd"}, 1},
		{"namespace of the root", `<notes xmlns="urn:notes"/>`, []string{"urn:notes", ""}, 1},
		{"nothing", `<notes/>`, []string{"", ""}, 0},
	}
	for _, tt := range tests {
		m, _ := newModel(t, tt.text)
		press(m, shiftV)
		if got := fieldTexts(m); !slices.Equal(got, tt.want) || m.schemaPopup.focused != tt.focused {
			t.Errorf("%s: fields %q, focused %d, want %q and %d", tt.name, got, m.schemaPopup.focused, tt.want, tt.focused)
		}
	}
}

func TestSaveSchemas(t *testing.T) {
	m, path := withNotesSchema(t, badNotes)
	if !strings.Contains(m.messageLine(), "Saved schemas.toml") {
		t.Errorf("message line = %q, want the saved file", m.messageLine())
	}
	if !strings.Contains(m.statusBar(), "✗ 2 XSD errors") {
		t.Errorf("status bar = %q, want two errors", m.statusBar())
	}
	want := []config.Schema{{Namespace: "urn:notes", Path: "notes.xsd"}}
	stored, err := config.LoadSchemas(filepath.Join(filepath.Dir(path), "schemas.toml"), path)
	if err != nil || !slices.Equal(stored, want) {
		t.Errorf("stored %v, %v, want %v", stored, err, want)
	}

	// The next start knows the schema.
	doc, err := document.Open(path, xml.Type())
	if err != nil {
		t.Fatal(err)
	}
	next := New(doc, setup(filepath.Dir(path), theme.Theme{}))
	settle(next, next.Init())
	if got := doc.Schemas(); !slices.Equal(got, want) || len(next.editor.Problems()) != 2 {
		t.Errorf("after a restart: schemas %v, %d problems, want %v and 2", got, len(next.editor.Problems()), want)
	}
}

func TestSchemaThatDoesNotLoadKeepsThePopupOpen(t *testing.T) {
	m, path := newModel(t, badNotes)
	writeNext(t, path, "notes.xsd", notesSchema)
	press(m, shiftV)
	press(m, typing("missing.xsd")...)
	press(m, enter)
	if m.schemaPopup == nil || !strings.Contains(m.schemaBox(), "there is no file") {
		t.Fatal("the popup closed without a usable schema")
	}
	press(m, backspace)
	if m.schemaPopup.err != nil {
		t.Error("the error stays after an edit")
	}

	press(m, esc)
	if m.schemaPopup != nil || m.editor.Document().Schemas() != nil {
		t.Error("esc did not close the popup without schemas")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "schemas.toml")); err == nil {
		t.Error("the schemas file was written")
	}
}

func TestNamespaceWithoutFile(t *testing.T) {
	m, _ := newModel(t, `<notes xmlns="urn:notes"/>`)
	press(m, shiftV, enter)
	if m.schemaPopup == nil || !strings.Contains(m.schemaBox(), "enter the schema file for urn:notes") {
		t.Error("the popup closed with a namespace without a file")
	}
}

func TestSchemaPopupEditsSeveralSchemas(t *testing.T) {
	m, _ := newModel(t, `<notes/>`)
	press(m, shiftV)
	press(m, typing("urn:a")...)
	press(m, down)
	press(m, typing("a.xsd")...)
	press(m, ctrlN)
	press(m, typing("urn:b")...)
	press(m, down)
	press(m, typing("b.xsd")...)
	if got, want := fieldTexts(m), []string{"urn:a", "a.xsd", "urn:b", "b.xsd"}; !slices.Equal(got, want) {
		t.Fatalf("fields %q, want %q", got, want)
	}
	view := m.schemaBox()
	if !strings.Contains(view, "Namespace  urn:b") || !strings.Contains(view, "▸ File       b.xsd") {
		t.Errorf("popup:\n%s", view)
	}

	press(m, up, up, ctrlD)
	if got, want := fieldTexts(m), []string{"urn:b", "b.xsd"}; !slices.Equal(got, want) || m.schemaPopup.focused != 0 {
		t.Errorf("after removing the first schema: %q, focused %d", got, m.schemaPopup.focused)
	}
	press(m, ctrlD)
	if got := fieldTexts(m); !slices.Equal(got, []string{"", ""}) {
		t.Errorf("after removing the last schema: %q, want it empty", got)
	}
}

func TestRemovingAllSchemas(t *testing.T) {
	m, path := withNotesSchema(t, badNotes)
	press(m, shiftV, ctrlD)
	check := press(m, enter)
	if len(m.editor.Problems()) != 0 {
		t.Error("the problems against the old schemas stay")
	}
	settle(m, check)
	if m.editor.Document().Schemas() != nil || len(m.editor.Problems()) != 0 || !strings.Contains(m.statusBar(), "✓ valid") {
		t.Errorf("schemas %v, problems %v, status bar %q, want none", m.editor.Document().Schemas(), m.editor.Problems(), m.statusBar())
	}
	if stored, _ := config.LoadSchemas(filepath.Join(filepath.Dir(path), "schemas.toml"), path); stored != nil {
		t.Errorf("the schemas file still has %v", stored)
	}
}

func TestSchemaPopupCursor(t *testing.T) {
	m, _ := newModel(t, `<notes/>`)
	press(m, shiftV, ctrlN)
	press(m, typing("abc")...)
	press(m, left)
	box := m.schemaBox()
	x, y := m.boxPosition(box)
	lines := strings.Split(box, "\n")
	c := m.cursor()
	line := []rune(lines[c.Y-y])
	// The cursor is on the c of abc.
	if got := string(line[c.X-x:]); !strings.HasPrefix(got, "c ") {
		t.Errorf("cursor at %q of\n%s", got, box)
	}
}

func TestPasteIntoSchemaPopup(t *testing.T) {
	m, _ := newModel(t, `<notes/>`)
	press(m, shiftV, down)
	m.Update(tea.PasteMsg{Content: "a.xsd\nmore"})
	if got := fieldTexts(m); got[1] != "a.xsd" {
		t.Errorf("fields after paste: %q", got)
	}
}

func TestCompletePath(t *testing.T) {
	m, path := newModel(t, `<notes/>`)
	for _, name := range []string{"books.xsd", "books-2.xsd", "notes.txt", ".hidden.xsd", "schemas/inner.xsd", "schemas/other.xsd"} {
		writeNext(t, path, name, notesSchema)
	}
	resolve := m.editor.Document().Resolve
	tests := []struct{ path, want string }{
		{"b", "books"},
		{"books-", "books-2.xsd"},
		{"s", "schemas/"},
		{"schemas/i", "schemas/inner.xsd"},
		{"schemas/", "schemas/"},
		{"n", "n"},
		{".", ".hidden.xsd"},
		{"missing/x", "missing/x"},
	}
	for _, tt := range tests {
		if got := completePath(tt.path, resolve); got != tt.want {
			t.Errorf("completePath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}

	press(m, shiftV, down)
	press(m, typing("sch")...)
	press(m, tab)
	press(m, typing("o")...)
	press(m, tab, up)
	press(m, typing("sch")...)
	press(m, tab)
	if got, want := fieldTexts(m), []string{"sch", "schemas/other.xsd"}; !slices.Equal(got, want) {
		t.Errorf("fields after tab = %q, want %q", got, want)
	}
}

func TestTextField(t *testing.T) {
	f := newTextField("ab")
	for _, k := range []tea.KeyPressMsg{left, key("x"), {Code: tea.KeyHome}, {Code: tea.KeyDelete}, {Code: tea.KeyEnd}, backspace, right, key("é")} {
		f.key(k)
	}
	if got := f.String(); got != "xé" || f.cursor != 2 {
		t.Errorf("text %q with the cursor at %d, want xé and 2", got, f.cursor)
	}
	f.set("abcdef")
	f.cursor = 1
	if text, col := f.view(3); text != "abc" || col != 1 {
		t.Errorf("view(3) = %q, %d at the start", text, col)
	}
	f.cursor = 6
	if text, col := f.view(3); text != "ef" || col != 2 {
		t.Errorf("view(3) = %q, %d at the end", text, col)
	}
}

func TestVWithoutSchemaOpensThePopup(t *testing.T) {
	m, _ := newModel(t, `<notes/>`)
	press(m, key("v"))
	if m.schemaPopup == nil {
		t.Error("v without a schema did not open the schema popup")
	}
}

func TestProblemList(t *testing.T) {
	m, _ := withNotesSchema(t, badNotes)
	m.editor.MoveTo(buffer.Position{Row: 3})
	press(m, key("v"))
	if m.problemList == nil || m.editor.Cursor() != (buffer.Position{Row: 4, Col: 3}) {
		t.Fatalf("v: list %v, cursor %v, want the problem after the cursor", m.problemList, m.editor.Cursor())
	}
	if got := m.messageLine(); !strings.Contains(got, "2/2") {
		t.Errorf("message line = %q, want the count", got)
	}
	if m.listHeight() != 3 || !strings.Contains(m.listView(3), `5  unexpected child element "{urn:notes}done"`) {
		t.Errorf("list:\n%s", m.listView(m.listHeight()))
	}

	press(m, down)
	if m.editor.Cursor() != (buffer.Position{Row: 2, Col: 3}) {
		t.Errorf("down: cursor %v, want the first problem", m.editor.Cursor())
	}
	press(m, enter)
	if m.problemList != nil || m.editor.Cursor().Row != 2 {
		t.Errorf("enter: list %v, cursor %v, want it closed at the problem", m.problemList, m.editor.Cursor())
	}
	if got := m.messageLine(); !strings.Contains(got, `unexpected child element "{urn:notes}todo"`) {
		t.Errorf("message line on the problem = %q", got)
	}

	press(m, key("v"), up, esc)
	if m.problemList != nil || m.editor.Cursor() != (buffer.Position{Row: 2, Col: 3}) {
		t.Errorf("esc: cursor %v, want it back", m.editor.Cursor())
	}
}

func TestValidAgainstTheSchema(t *testing.T) {
	m, _ := withNotesSchema(t, "<notes xmlns=\"urn:notes\">\n  <note>a</note>\n</notes>")
	if !strings.Contains(m.statusBar(), "✓ XSD valid") {
		t.Errorf("status bar = %q", m.statusBar())
	}
	press(m, key("v"))
	if m.problemList != nil || !strings.Contains(m.messageLine(), "No errors against the schema") {
		t.Errorf("v: list %v, message line %q", m.problemList, m.messageLine())
	}
}

func TestProblemsFollowEditsUntilTheNextCheck(t *testing.T) {
	m, _ := withNotesSchema(t, badNotes)
	m.editor.MoveTo(buffer.Position{Row: 1})
	press(m, key("i"), enter, enter)
	if rows := []int{m.editor.Problems()[0].Range.Start.Row, m.editor.Problems()[1].Range.Start.Row}; !slices.Equal(rows, []int{4, 6}) {
		t.Errorf("problems in rows %v after two new lines, want 4 and 6", rows)
	}
	m.editor.MoveTo(buffer.Position{Row: 4, Col: 2})
	press(m, typing("<note/>")...)
	press(m, tea.KeyPressMsg{Code: tea.KeyDelete})
	for range len("<todo/>") {
		press(m, tea.KeyPressMsg{Code: tea.KeyDelete})
	}
	settle(m, press(m, esc))
	if !strings.Contains(m.statusBar(), "✗ 1 XSD error") {
		t.Errorf("status bar = %q, want one error left", m.statusBar())
	}
}

func TestSyntaxErrorWithSchema(t *testing.T) {
	m, _ := withNotesSchema(t, "<notes xmlns=\"urn:notes\">\n  <note>a</nope>\n</notes>")
	if !strings.Contains(m.statusBar(), "✗ line 2") {
		t.Errorf("status bar = %q, want the line of the syntax error", m.statusBar())
	}
	problems := m.editor.Problems()
	if len(problems) != 1 || problems[0].Range != (buffer.Range{Start: buffer.Position{Row: 1, Col: 2}, End: buffer.Position{Row: 1, Col: 16}}) {
		t.Errorf("problems = %v, want the line of the syntax error", problems)
	}
}

func TestStoredSchemaThatDoesNotLoad(t *testing.T) {
	dir := t.TempDir()
	path := writeNext(t, filepath.Join(dir, "x"), "test.xml", badNotes)
	if err := config.SaveSchemas(filepath.Join(dir, "schemas.toml"), path, []config.Schema{{Namespace: "urn:notes", Path: "gone.xsd"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := document.Open(path, xml.Type())
	if err != nil {
		t.Fatal(err)
	}
	m := New(doc, setup(dir, theme.Theme{}))
	settle(m, m.Init())
	if !strings.Contains(m.statusBar(), "✗ schema") || !strings.Contains(m.messageLine(), "Cannot load the schema: there is no file") {
		t.Errorf("status bar %q, message line %q, want the schema error", m.statusBar(), m.messageLine())
	}
}

func TestChangedSchemaIsCheckedAgain(t *testing.T) {
	m, path := withNotesSchema(t, badNotes)
	writeNext(t, path, "notes.xsd", strings.Replace(notesSchema, `<xs:element name="note" type="xs:string" maxOccurs="unbounded"/>`, `<xs:any processContents="skip" maxOccurs="unbounded"/>`, 1))
	cmd := press(m, key("v"))
	if !strings.Contains(m.messageLine(), "The schema changed") {
		t.Fatalf("message line = %q, want the change", m.messageLine())
	}
	settle(m, cmd)
	if !strings.Contains(m.statusBar(), "✓ XSD valid") {
		t.Errorf("status bar = %q, want the new schema used", m.statusBar())
	}
}

func TestManyProblems(t *testing.T) {
	m, _ := withNotesSchema(t, badNotes)
	m.editor.SetProblems(make([]filetype.Problem, filetype.MaxProblems))
	if !strings.Contains(m.statusBar(), "✗ 10000+ XSD errors") {
		t.Errorf("status bar = %q", m.statusBar())
	}
}

func TestFileTypeWithoutSchemas(t *testing.T) {
	doc := document.New("", filetype.Plain())
	m := New(doc, setup(t.TempDir(), theme.Theme{}))
	press(m, key("v"))
	if m.schemaPopup != nil || !strings.Contains(m.messageLine(), "Text files have no schemas") {
		t.Errorf("v: popup %v, message line %q", m.schemaPopup, m.messageLine())
	}
	press(m, shiftV)
	if m.schemaPopup != nil {
		t.Error("V opened the popup")
	}
}

func TestCheckWithOldSchemasIsDropped(t *testing.T) {
	m, _ := withNotesSchema(t, badNotes)
	press(m, key("i"), key("x"))
	check := press(m, esc)
	if check == nil {
		t.Fatal("esc did not check the changed text")
	}
	// While the text is checked against notes.xsd, the schema goes.
	press(m, shiftV, ctrlD, enter)
	next := settleOne(m, check)
	if len(m.editor.Problems()) != 0 {
		t.Errorf("the check against the old schema shows: %v", m.editor.Problems())
	}
	if next == nil {
		t.Fatal("the old result did not start a check with the new schemas")
	}
	settle(m, next)
	if !strings.Contains(m.statusBar(), "✗ line 1") {
		t.Errorf("status bar = %q, want the text without schema", m.statusBar())
	}
}

// settleOne runs the check in cmd, which may be one of a batch, hands its
// result to m and returns the command m answers with.
func settleOne(m *Model, cmd tea.Cmd) tea.Cmd {
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			if result, ok := c().(validatedMsg); ok {
				msg = result
			}
		}
	}
	_, next := m.Update(msg)
	return next
}
