package xml

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/sample"
)

func schemaFile(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", "schemas", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func loadSchemas(t *testing.T, schemas ...config.Schema) filetype.SchemaCheck {
	t.Helper()
	for i := range schemas {
		schemas[i].Path = schemaFile(t, schemas[i].Path)
	}
	check, err := Schemas{}.Load(schemas)
	if err != nil {
		t.Fatal(err)
	}
	return check
}

func books(t *testing.T) filetype.SchemaCheck {
	return loadSchemas(t, config.Schema{Namespace: "urn:books", Path: "books.xsd"})
}

// problemsIn shows each problem as the text it marks and its message.
func problemsIn(src string, problems []filetype.Problem) []string {
	lines := strings.Split(src, "\n")
	var out []string
	for _, p := range problems {
		line := []rune(lines[p.Range.Start.Row])
		out = append(out, fmt.Sprintf("%d %q: %s", p.Range.Start.Row+1, string(line[p.Range.Start.Col:p.Range.End.Col]), p.Message))
	}
	return out
}

func TestSchemaCheckFindsAllProblems(t *testing.T) {
	src := `<?xml version="1.0"?>
<library xmlns="urn:books" xmlns:m="urn:meta">
  <book id="b1" lang="de">
    <title>Über</title><price>abc</price>
  </book>
  <book lang="de"><title>Über das Buch</title><price>1</price><m:note>x</m:note><extra/></book>
  <book id="b3">
    <price>2</price>
  </book>
</library>`
	got := problemsIn(src, books(t)(src))
	want := []string{
		`4 "abc": invalid simple content: value "abc" is not valid for xs:decimal: invalid decimal`,
		`6 "book": missing required attribute "id"`,
		`6 "extra": unexpected child element "{urn:books}extra"; expected next child [an element in namespace ["urn:extra"]]`,
		`8 "price": unexpected child element "{urn:books}price"; expected next child ["{urn:books}title"]`,
		`9 "book": content is incomplete; expected next child ["{urn:books}title"]`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("problems:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestSchemaCheckOfValidDocument(t *testing.T) {
	src := `<library xmlns="urn:books" xmlns:m="urn:meta">
  <book id="b1"><title>Ü</title><price>1.5</price><m:note>from the import</m:note></book>
</library>`
	if problems := books(t)(src); len(problems) != 0 {
		t.Errorf("problems in a valid document: %v", problems)
	}
}

func TestSchemasOfSeveralNamespaces(t *testing.T) {
	src := `<library xmlns="urn:books"><book id="b1"><title/><price>1</price><stock xmlns="urn:extra">-1</stock></book></library>`
	if problems := books(t)(src); len(problems) != 1 || !strings.Contains(problems[0].Message, "urn:extra") {
		t.Errorf("without the schema of urn:extra: %v, want its element unknown", problems)
	}
	check := loadSchemas(t,
		config.Schema{Namespace: "urn:books", Path: "books.xsd"},
		config.Schema{Namespace: "urn:extra", Path: "extra.xsd"},
	)
	got := problemsIn(src, check(src))
	if len(got) != 1 || !strings.Contains(got[0], `"-1"`) {
		t.Errorf("with both schemas: %v, want the negative stock", got)
	}
}

func TestSchemaWithoutNamespace(t *testing.T) {
	check := loadSchemas(t, config.Schema{Path: "notes.xsd"})
	src := "<notes>\n  <note>a</note>\n  <todo/>\n</notes>"
	if got := problemsIn(src, check(src)); len(got) != 1 || !strings.HasPrefix(got[0], `3 "todo"`) {
		t.Errorf("problems = %v, want the todo element", got)
	}
}

func TestLoadSchemaErrors(t *testing.T) {
	tests := []struct {
		name   string
		schema config.Schema
		want   string
	}{
		{"wrong namespace", config.Schema{Namespace: "urn:other", Path: "books.xsd"}, `books.xsd is for the namespace "urn:books", not for the namespace "urn:other"`},
		{"namespace missing", config.Schema{Path: "books.xsd"}, `books.xsd is for the namespace "urn:books", not for no namespace`},
		{"schema without namespace", config.Schema{Namespace: "urn:notes", Path: "notes.xsd"}, `notes.xsd is for no namespace, not for the namespace "urn:notes"`},
		{"missing file", config.Schema{Path: "missing.xsd"}, "there is no file "},
		{"no schema", config.Schema{Path: "../../xml.go"}, "xml.go is no XSD schema"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.schema.Path = schemaFile(t, tt.schema.Path)
			_, err := Schemas{}.Load([]config.Schema{tt.schema})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSchemaCheckFindsManyProblems(t *testing.T) {
	var b strings.Builder
	b.WriteString(`<library xmlns="urn:books">`)
	for i := range filetype.MaxProblems + 1 {
		fmt.Fprintf(&b, "\n<book id=\"b%d\"><title/><price>x</price></book>", i)
	}
	b.WriteString("\n</library>")
	problems := books(t)(b.String())
	if len(problems) != filetype.MaxProblems {
		t.Fatalf("found %d problems, want %d", len(problems), filetype.MaxProblems)
	}
	if last := problems[len(problems)-1].Range.Start; last.Row != filetype.MaxProblems {
		t.Errorf("last problem in row %d, want %d", last.Row, filetype.MaxProblems)
	}
}

func TestSchemaCheckWithDoctype(t *testing.T) {
	src := "<!DOCTYPE notes [\n  <!ELEMENT notes ANY>\n]><notes><todo/></notes>"
	check := loadSchemas(t, config.Schema{Path: "notes.xsd"})
	got := problemsIn(src, check(src))
	if len(got) != 2 || !strings.HasPrefix(got[0], `3 "todo"`) || !strings.HasPrefix(got[1], `3 "notes"`) {
		t.Errorf("problems = %v, want the todo element and the end of notes", got)
	}
}

func TestMarked(t *testing.T) {
	tests := []struct {
		line  string
		col   int
		value bool
		want  string
	}{
		{"  <book id='1'>", 2, false, "book"},
		{"<a><b>x</b></a>", 7, false, "b"},
		{"<price> 12a </price>", 12, true, "12a"},
		{"    </price>", 4, true, "price"},
		{"<p:price>1</p:price>", 0, false, "p:price"},
		{"text here", 0, false, "text"},
		{"<>", 0, false, "<"},
		{`<a><b id="1"/>`, 13, false, "b"},
		{`<a><b id="1"/>`, 12, false, "b"},
		{"a > b", 2, false, ">"},
		{"ab", 5, false, "b"},
		{"", 0, false, ""},
	}
	for _, tt := range tests {
		line := []rune(tt.line)
		from, to := marked(line, tt.col, tt.value)
		if got := string(line[from:to]); got != tt.want {
			t.Errorf("marked(%q, %d) = %q, want %q", tt.line, tt.col, got, tt.want)
		}
	}
}

func TestNamedSchemas(t *testing.T) {
	tests := []struct {
		name, src string
		want      []config.Schema
	}{
		{"schema location", `<?xml version="1.0"?><!-- x --><a xmlns="urn:a" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="urn:a a.xsd
			urn:b  b.xsd"/>`, []config.Schema{{Namespace: "urn:a", Path: "a.xsd"}, {Namespace: "urn:b", Path: "b.xsd"}}},
		{"without namespace", `<a xmlns:i="http://www.w3.org/2001/XMLSchema-instance" i:noNamespaceSchemaLocation=" a.xsd "/>`, []config.Schema{{Path: "a.xsd"}}},
		{"namespace of the root", `<p:a xmlns:p="urn:a"><b/></p:a>`, []config.Schema{{Namespace: "urn:a"}}},
		{"no root", `<?xml version="1.0"?>`, nil},
	}
	for _, tt := range tests {
		if got := (Schemas{}).Named(tt.src); !slices.Equal(got, tt.want) {
			t.Errorf("%s: Named() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestProblemsAreSortedByPosition(t *testing.T) {
	// The validator finds the missing key after the duplicate one, when the
	// element of the keys ends.
	src := `<lib><ref to="9"/><b id="1"/><b id="1"/></lib>`
	got := problemsIn(src, loadSchemas(t, config.Schema{Path: "keys.xsd"})(src))
	if len(got) != 2 || !strings.HasPrefix(got[0], `1 "ref": keyref`) || !strings.HasPrefix(got[1], `1 "b": duplicate`) {
		t.Errorf("problems = %q, want the missing key first", got)
	}
}

func TestSampleIsValidAgainstItsSchema(t *testing.T) {
	src := sample.XML(100_000)
	if problems := loadSchemas(t, config.Schema{Path: "library.xsd"})(src); len(problems) != 0 {
		t.Errorf("problems in the sample: %v", problems)
	}
}

func TestSchemaCheckOfLargeDocuments(t *testing.T) {
	check := loadSchemas(t, config.Schema{Path: "notes.xsd"})
	long := "<notes><note>" + strings.Repeat("x", 5<<20) + "</note></notes>"
	if problems := check(long); len(problems) != 0 {
		t.Errorf("problems in a long text: %v", problems[0])
	}
	if testing.Short() {
		t.Skip("large document")
	}
	large := sample.XML(65 << 20)
	if problems := loadSchemas(t, config.Schema{Path: "library.xsd"})(large); len(problems) != 0 {
		t.Errorf("problems in a document of 65 MB: %v", problems[0])
	}
}
