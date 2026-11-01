package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/buffer"
)

func TestFarParentIsFoundInTheBackground(t *testing.T) {
	m, _ := newModel(t, "<a>\n"+strings.Repeat("<b/>\n", 10_000)+"</a>")
	press(m, tea.KeyPressMsg{Code: tea.KeyEnd, Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: tea.KeyHome})
	cmd := press(m, tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModCtrl})
	if !strings.Contains(m.messageLine(), jumpHint) {
		t.Errorf("message line = %q, want the hint while the job looks", m.messageLine())
	}
	if press(m, tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModCtrl}) != nil {
		t.Error("a second job started while the first one looks")
	}
	settle(m, cmd)
	if got := m.editor.Cursor(); got != (buffer.Position{}) {
		t.Errorf("cursor = %v, want the start tag of the root", got)
	}
	if strings.Contains(m.messageLine(), jumpHint) {
		t.Error("the hint stays after the jump")
	}
}
