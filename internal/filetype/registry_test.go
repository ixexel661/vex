package filetype

import "testing"

func TestDetect(t *testing.T) {
	r := NewRegistry(Type{Name: "XML", Extensions: []string{".xml"}})
	r.Register(Type{Name: "JSON", Extensions: []string{".json"}})

	tests := map[string]string{
		"dir/test.xml":  "XML",
		"TEST.XML":      "XML",
		"data.json":     "JSON",
		"notes.txt":     "Text",
		"no-extension":  "Text",
		"archive.xml.d": "Text",
	}
	for path, want := range tests {
		if got := r.Detect(path).Name; got != want {
			t.Errorf("Detect(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestDefault(t *testing.T) {
	if got := NewRegistry().Default().Name; got != "Text" {
		t.Errorf("Default() of an empty registry = %q, want Text", got)
	}
	r := NewRegistry(Type{Name: "XML"}, Type{Name: "JSON"})
	if got := r.Default().Name; got != "XML" {
		t.Errorf("Default() = %q, want the first type, XML", got)
	}
}
