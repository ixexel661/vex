package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/sample"
	"github.com/ixexel661/vex/internal/theme"
)

func TestValidationRunsInTheBackground(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.xml")
	if err := os.WriteFile(path, []byte("<a/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := document.Open(path, xml.Type())
	if err != nil {
		t.Fatal(err)
	}
	m := New(doc, setup(dir, theme.Theme{}))
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init did not start the check")
	}
	if !strings.Contains(m.statusBar(), "… checking") {
		t.Errorf("status bar = %q, want it checking", m.statusBar())
	}
	settle(m, cmd)
	if !strings.Contains(m.statusBar(), "✓ valid") {
		t.Errorf("status bar = %q, want the result", m.statusBar())
	}
	if m.prefetching || m.prefetch() != nil {
		t.Error("the highlighting is not complete after the start")
	}
}

func TestValidationOfOldTextIsDropped(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	press(m, key("i"), key("x"))
	_, check := m.Update(esc)
	if check == nil {
		t.Fatal("esc did not check the changed text")
	}
	// While "x<a/>" is checked, the x goes again.
	press(m, key("i"), backspace)
	if _, cmd := m.Update(esc); cmd != nil {
		t.Fatal("a second check started while the first one runs")
	}

	_, next := m.Update(check())
	if next == nil {
		t.Fatal("the result for the old text did not start a check of the new one")
	}
	if m.validation.err != nil || strings.Contains(m.messageLine(), "Invalid") {
		t.Errorf("the result for the old text shows: %v, %q", m.validation.err, m.messageLine())
	}
	settle(m, next)
	if m.validation.err != nil || !strings.Contains(m.statusBar(), "✓ valid") {
		t.Errorf("err = %v, status bar = %q, want the new text valid", m.validation.err, m.statusBar())
	}
}

func TestOnePrefetchAtATime(t *testing.T) {
	m, _ := newModel(t, sample.XML(20_000))
	m.editor.SetText(sample.XML(30_000))
	if m.prefetch() == nil {
		t.Fatal("no job for the new text")
	}
	if m.prefetch() != nil {
		t.Error("a second job started while the first one runs")
	}
}
