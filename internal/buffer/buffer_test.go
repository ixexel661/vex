package buffer

import "testing"

func TestTextRoundTrip(t *testing.T) {
	text := "<a>\n  <b/>\n</a>\n"
	b := New(text)
	if got := b.Text(); got != text {
		t.Errorf("Text() = %q, want %q", got, text)
	}
	if got := b.LineCount(); got != 4 {
		t.Errorf("LineCount() = %d, want 4", got)
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		at      Position
		insert  string
		want    string
		wantEnd Position
	}{
		{"within line", "hello", Position{0, 5}, " world", "hello world", Position{0, 11}},
		{"multiple lines", "ab", Position{0, 1}, "x\ny", "ax\nyb", Position{1, 1}},
		{"newline", "ab", Position{0, 1}, "\n", "a\nb", Position{1, 0}},
		{"unicode", "äöü", Position{0, 1}, "x", "äxöü", Position{0, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.text)
			end := b.Insert(tt.at, tt.insert)
			if got := b.Text(); got != tt.want {
				t.Errorf("Text() = %q, want %q", got, tt.want)
			}
			if end != tt.wantEnd {
				t.Errorf("Insert() = %v, want %v", end, tt.wantEnd)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		from, to Position
		want     string
	}{
		{"within line", "abcd", Position{0, 1}, Position{0, 3}, "ad"},
		{"across lines", "abc\ndef", Position{0, 1}, Position{1, 2}, "af"},
		{"line break", "ab\ncd", Position{0, 2}, Position{1, 0}, "abcd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.text)
			b.Delete(tt.from, tt.to)
			if got := b.Text(); got != tt.want {
				t.Errorf("Text() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmptyDeleteKeepsVersion(t *testing.T) {
	b := New("abc")
	b.Delete(Position{0, 1}, Position{0, 1})
	if b.Version() != 0 {
		t.Errorf("Version() = %d, want 0", b.Version())
	}
}

func TestPrevNextWrapLines(t *testing.T) {
	b := New("ab\ncd")
	if got := b.Next(Position{0, 2}); got != (Position{1, 0}) {
		t.Errorf("Next() = %v, want {1 0}", got)
	}
	if got := b.Prev(Position{1, 0}); got != (Position{0, 2}) {
		t.Errorf("Prev() = %v, want {0 2}", got)
	}
	if got := b.Prev(Position{0, 0}); got != (Position{0, 0}) {
		t.Errorf("Prev() at start = %v, want {0 0}", got)
	}
	if got := b.Next(Position{1, 2}); got != (Position{1, 2}) {
		t.Errorf("Next() at end = %v, want {1 2}", got)
	}
}

func TestClamp(t *testing.T) {
	b := New("abc\nd")
	if got := b.Clamp(Position{5, 9}); got != (Position{1, 1}) {
		t.Errorf("Clamp() = %v, want {1 1}", got)
	}
	if got := b.Clamp(Position{-1, -1}); got != (Position{0, 0}) {
		t.Errorf("Clamp() = %v, want {0 0}", got)
	}
}
