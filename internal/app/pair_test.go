package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestFarPairIsFoundInTheBackground(t *testing.T) {
	m, _ := newModel(t, "<a>\n"+strings.Repeat("<b/>\n", 10_000)+"</a>")
	press(m, key("i"), tea.KeyPressMsg{Code: tea.KeyEnd, Mod: tea.ModCtrl})
	cmd := press(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	if !m.findingPair {
		t.Fatal("no job looks for the start tag of the root")
	}
	if press(m, tea.KeyPressMsg{Code: tea.KeyLeft}) != nil {
		t.Error("a second job started while the first one runs")
	}
	settle(m, cmd)
	press(m, key("x"))
	text := m.editor.Document().Buffer.Text()
	if !strings.HasPrefix(text, "<xa>") || !strings.HasSuffix(text, "</xa>") {
		t.Errorf("typing renamed %q … %q, want both tags", text[:5], text[len(text)-6:])
	}
}
