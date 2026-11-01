package printable

import "testing"

func TestString(t *testing.T) {
	tests := []struct{ in, want string }{
		{"plain – Über", "plain – Über"},
		{"a\tb", "a b"},
		{"a\x1b]52;c;dGVzdA==\a.xml", "a?]52;c;dGVzdA==?.xml"},
		{"\x00\x7f\u0085\u009b", "????"},
		{"a\x9bb", "a�b"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := String(tt.in); got != tt.want {
			t.Errorf("String(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
