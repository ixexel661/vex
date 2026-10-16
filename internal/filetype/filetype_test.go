package filetype

import "testing"

func TestFileName(t *testing.T) {
	xml := Type{Extensions: []string{".xml", ".svg"}}
	tests := []struct {
		fileType Type
		name     string
		want     string
	}{
		{xml, "test", "test.xml"},
		{xml, "dir.d/test", "dir.d/test.xml"},
		{xml, "test.xml", "test.xml"},
		{xml, "notes.txt", "notes.txt"},
		{Plain(), "test", "test"},
	}
	for _, tt := range tests {
		if got := tt.fileType.FileName(tt.name); got != tt.want {
			t.Errorf("FileName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
