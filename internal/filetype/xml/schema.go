package xml

import (
	stdxml "encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jacoelho/xsd"
	"github.com/jacoelho/xsd/xsderrors"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
)

const (
	xsdNamespace = "http://www.w3.org/2001/XMLSchema"
	xsiNamespace = "http://www.w3.org/2001/XMLSchema-instance"
)

// Schemas checks documents against XSD files.
type Schemas struct{}

func (Schemas) Name() string {
	return "XSD"
}

// Load compiles the XSD files, after it made sure each one is for the
// namespace it is given for. Their includes and imports come along.
func (Schemas) Load(schemas []config.Schema) (filetype.SchemaCheck, error) {
	sources := make([]xsd.SchemaSource, len(schemas))
	for i, s := range schemas {
		namespace, err := targetNamespace(s.Path)
		if err != nil {
			return nil, err
		}
		if namespace != s.Namespace {
			return nil, fmt.Errorf("%s is for %s, not for %s", filepath.Base(s.Path), namespaceName(namespace), namespaceName(s.Namespace))
		}
		sources[i] = xsd.File(s.Path)
	}
	engine, err := xsd.Compile(sources...)
	if err != nil {
		return nil, schemaError(err)
	}
	return func(src string) []filetype.Problem {
		return check(engine, src)
	}, nil
}

// targetNamespace reads the namespace an XSD file is for.
func targetNamespace(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", readError(path, err)
	}
	defer f.Close()
	d := stdxml.NewDecoder(f)
	d.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return "", fmt.Errorf("%s is no XSD schema: %w", filepath.Base(path), toSyntaxError(err))
		}
		if start, ok := tok.(stdxml.StartElement); ok {
			if start.Name != (stdxml.Name{Space: xsdNamespace, Local: "schema"}) {
				return "", fmt.Errorf("%s is no XSD schema", filepath.Base(path))
			}
			return attr(start, stdxml.Name{Local: "targetNamespace"}), nil
		}
	}
}

func readError(path string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("there is no file %s", path)
	}
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}
	return fmt.Errorf("cannot read %s: %w", path, err)
}

func namespaceName(namespace string) string {
	if namespace == "" {
		return "no namespace"
	}
	return fmt.Sprintf("the namespace %q", namespace)
}

func attr(start stdxml.StartElement, name stdxml.Name) string {
	for _, a := range start.Attr {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

// schemaError is the first problem of a schema that does not compile.
func schemaError(err error) error {
	for _, e := range xsderrors.Flatten(err) {
		var d *xsderrors.Error
		if !errors.As(e, &d) {
			continue
		}
		message := diagnosticMessage(d)
		if d.Line() > 0 {
			return fmt.Errorf("%s:%d: %s", filepath.Base(d.Path()), d.Line(), message)
		}
		return errors.New(message)
	}
	return err
}

// diagnosticMessage is the message of d, which some leave empty.
func diagnosticMessage(d *xsderrors.Error) string {
	if cause := d.Cause(); cause != nil && d.Code() == xsderrors.CodeSchemaRead {
		var pathErr *fs.PathError
		if errors.As(cause, &pathErr) {
			return readError(pathErr.Path, pathErr.Err).Error()
		}
	}
	if d.Message() != "" {
		return d.Message()
	}
	switch d.Code() {
	case xsderrors.CodeUnsupportedDTD:
		return "DTDs are not supported"
	case xsderrors.CodeUnsupportedEntity, xsderrors.CodeUnsupportedExternal:
		return "entities declared in a DTD are not supported"
	}
	return d.Error()
}

// check validates src against the compiled schemas. The limits grow with
// the document, so large files are checked to the end.
func check(engine *xsd.Engine, src string) []filetype.Problem {
	size := int64(len(src)) + 1
	err := engine.ValidateWithOptions(withoutDoctype(src), xsd.ValidateOptions{
		MaxErrors:             filetype.MaxProblems,
		MaxInstanceBytes:      size,
		MaxInstanceTextBytes:  size,
		MaxInstanceTokenBytes: size,
		MaxIdentityEntries:    int(size),
	})
	var diagnostics []*xsderrors.Error
	var problems []filetype.Problem
	for _, e := range xsderrors.Flatten(err) {
		var d *xsderrors.Error
		if errors.As(e, &d) {
			diagnostics = append(diagnostics, d)
		} else {
			problems = append(problems, filetype.Problem{Message: e.Error()})
		}
	}
	return slices.SortedStableFunc(slices.Values(append(problems, problemsAt(src, diagnostics)...)), func(a, b filetype.Problem) int {
		return a.Range.Start.Compare(b.Range.Start)
	})
}

// withoutDoctype blanks out the document type declaration, which the
// validator refuses, so the lines and columns of the rest stay the same.
func withoutDoctype(src string) io.Reader {
	d := newDecoder(src)
	for {
		start := d.InputOffset()
		tok, err := d.RawToken()
		if err != nil {
			return strings.NewReader(src)
		}
		switch t := tok.(type) {
		case stdxml.StartElement:
			return strings.NewReader(src)
		case stdxml.Directive:
			if strings.HasPrefix(string(t), "DOCTYPE") {
				end := d.InputOffset()
				return io.MultiReader(
					strings.NewReader(src[:start]),
					strings.NewReader(blank(src[start:end])),
					strings.NewReader(src[end:]),
				)
			}
		}
	}
}

// blank replaces every byte but line breaks with a space.
func blank(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] != '\n' {
			b[i] = ' '
		}
	}
	return string(b)
}

// problemsAt turns diagnostics into problems at their place in src. It
// goes through src once, in the order of the lines.
func problemsAt(src string, diagnostics []*xsderrors.Error) []filetype.Problem {
	slices.SortStableFunc(diagnostics, func(a, b *xsderrors.Error) int {
		return a.Line() - b.Line()
	})
	problems := make([]filetype.Problem, len(diagnostics))
	row, start := 0, 0 // the row of the line that starts at byte start
	for i, d := range diagnostics {
		for row < d.Line()-1 {
			next := strings.IndexByte(src[start:], '\n')
			if next < 0 {
				break
			}
			start += next + 1
			row++
		}
		line := src[start:]
		if end := strings.IndexByte(line, '\n'); end >= 0 {
			line = line[:end]
		}
		col := utf8.RuneCountInString(line[:min(max(d.Column()-1, 0), len(line))])
		from, to := marked([]rune(line), col, isValueCode(d.Code()))
		problems[i] = filetype.Problem{
			Range:   buffer.Range{Start: buffer.Position{Row: row, Col: from}, End: buffer.Position{Row: row, Col: to}},
			Message: diagnosticMessage(d),
		}
	}
	return problems
}

// isValueCode reports whether a problem is about a value, which the
// validator reports at the end tag after it.
func isValueCode(code xsderrors.Code) bool {
	return code == xsderrors.CodeValidationFacet || code == xsderrors.CodeValidationText || code == xsderrors.CodeValidationType
}

// marked returns the runes of line from "from" to "to" that a problem at
// col is about: the name of the tag there, or for a wrong value the text
// before its end tag on the same line, otherwise the word at col.
func marked(line []rune, col int, value bool) (from, to int) {
	if col >= len(line) {
		return max(len(line)-1, 0), len(line)
	}
	if line[col] == '>' || line[col] == '/' {
		// The end of a tag stands for the tag.
		for start := col - 1; start >= 0; start-- {
			if line[start] == '<' {
				col = start
				break
			}
		}
	}
	if line[col] != '<' {
		return col, wordEnd(line, col)
	}
	name := col + 1
	if name < len(line) && line[name] == '/' {
		name++
		if value {
			if from, to, ok := textBefore(line, col); ok {
				return from, to
			}
		}
	}
	end := name
	for end < len(line) && isNameRune(line[end]) {
		end++
	}
	if end == name {
		return col, col + 1
	}
	return name, end
}

// textBefore returns the text between the tag that ends before col and
// col, without the spaces around it.
func textBefore(line []rune, col int) (from, to int, ok bool) {
	from = col
	for from > 0 && line[from-1] != '>' {
		from--
	}
	if from == 0 {
		return 0, 0, false
	}
	to = col
	for from < to && unicode.IsSpace(line[from]) {
		from++
	}
	for to > from && unicode.IsSpace(line[to-1]) {
		to--
	}
	return from, to, from < to
}

func wordEnd(line []rune, col int) int {
	end := col + 1
	for end < len(line) && !unicode.IsSpace(line[end]) && line[end] != '<' && line[end] != '>' {
		end++
	}
	return end
}

// Named reads xsi:schemaLocation and xsi:noNamespaceSchemaLocation of the
// root element.
func (Schemas) Named(src string) []config.Schema {
	d := newDecoder(src)
	for {
		tok, err := d.Token()
		if err != nil {
			return nil
		}
		root, ok := tok.(stdxml.StartElement)
		if !ok {
			continue
		}
		var named []config.Schema
		for _, a := range root.Attr {
			if a.Name.Space != xsiNamespace {
				continue
			}
			switch a.Name.Local {
			case "schemaLocation":
				fields := strings.Fields(a.Value)
				for i := 0; i+1 < len(fields); i += 2 {
					named = append(named, config.Schema{Namespace: fields[i], Path: fields[i+1]})
				}
			case "noNamespaceSchemaLocation":
				named = append(named, config.Schema{Path: strings.TrimSpace(a.Value)})
			}
		}
		if len(named) == 0 {
			named = []config.Schema{{Namespace: root.Name.Space}}
		}
		return named
	}
}
