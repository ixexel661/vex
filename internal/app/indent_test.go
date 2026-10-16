package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

var shiftTab = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}

func TestTabKeysIndent(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	press(m, key("i"), tab)
	if got := m.editor.Document().Buffer.Text(); got != "  <a/>" {
		t.Errorf("text after tab = %q, want %q", got, "  <a/>")
	}
	press(m, shiftTab)
	if got := m.editor.Document().Buffer.Text(); got != "<a/>" {
		t.Errorf("text after shift+tab = %q, want %q", got, "<a/>")
	}
}

func TestTabFollowsIndentSetting(t *testing.T) {
	m, _ := newModel(t, "<a/>")
	// The new indent formats the document too, which ends it with a newline.
	press(m, key("s"), down, down, right, enter, key("i"), tab)
	if got := m.editor.Document().Buffer.Text(); got != "    <a/>\n" {
		t.Errorf("text = %q, want the indent of 4 spaces from the settings", got)
	}
}
