package command

import (
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/editor"
	"github.com/ixexel661/vex/internal/mode"
)

// Context is everything a command can use of the running application.
// It is also the API a plugin system would build on.
type Context interface {
	Editor() *editor.Editor
	Config() config.Config
	SetMode(m mode.Mode)
	Notify(message string)
	Ask(question string, choices ...Choice)
	OpenSearch()
	OpenSettings()
	Quit()
}

// Command is a named action that keys can be bound to. Commands with a
// Title are advertised in the key hints.
type Command struct {
	Name  string
	Title string
	Run   func(ctx Context) error
}

// Choice is a possible answer to a question asked with Context.Ask.
type Choice struct {
	Key   string
	Label string
	Run   func(ctx Context) error
}
