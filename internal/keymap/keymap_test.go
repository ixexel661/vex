package keymap

import (
	"testing"

	"github.com/ixexel661/vex/internal/mode"
)

func TestBindReplacesExistingKey(t *testing.T) {
	k := New()
	k.Bind(mode.Normal, "x", "first")
	k.Bind(mode.Normal, "x", "second")

	if got, _ := k.Lookup(mode.Normal, "x"); got != "second" {
		t.Errorf("Lookup() = %q, want %q", got, "second")
	}
	if got := len(k.Bindings(mode.Normal)); got != 1 {
		t.Errorf("got %d bindings, want 1", got)
	}
}

func TestBindingsArePerMode(t *testing.T) {
	k := Default()
	if got, _ := k.Lookup(mode.Normal, "q"); got != "app.quit" {
		t.Errorf("normal q = %q, want app.quit", got)
	}
	if _, ok := k.Lookup(mode.Insert, "q"); ok {
		t.Error("q is bound in insert mode, it must insert text there")
	}
}
