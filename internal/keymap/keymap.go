package keymap

import (
	"slices"

	"github.com/ixexel661/vex/internal/mode"
)

// Binding connects a key, as named by Bubble Tea (e.g. "ctrl+s"), to the
// name of a command.
type Binding struct {
	Key     string
	Command string
}

// Keymap holds the bindings of each mode in the order they were added.
type Keymap struct {
	bindings map[mode.Mode][]Binding
}

func New() *Keymap {
	return &Keymap{bindings: map[mode.Mode][]Binding{}}
}

// Bind maps key to command in mode and replaces an existing binding of key.
func (k *Keymap) Bind(m mode.Mode, key, command string) {
	for i, b := range k.bindings[m] {
		if b.Key == key {
			k.bindings[m][i].Command = command
			return
		}
	}
	k.bindings[m] = append(k.bindings[m], Binding{Key: key, Command: command})
}

func (k *Keymap) Lookup(m mode.Mode, key string) (string, bool) {
	for _, b := range k.bindings[m] {
		if b.Key == key {
			return b.Command, true
		}
	}
	return "", false
}

func (k *Keymap) Bindings(m mode.Mode) []Binding {
	return slices.Clone(k.bindings[m])
}
