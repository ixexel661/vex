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
