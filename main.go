package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/app"
	"github.com/ixexel661/vex/internal/command"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/filetype/xml"
	"github.com/ixexel661/vex/internal/keymap"
	"github.com/ixexel661/vex/internal/search"
	"github.com/ixexel661/vex/internal/theme"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: vex <file>")
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(flag.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, "vex:", err)
		os.Exit(1)
	}
}

// run wires everything together. All registries are created here, so a
// plugin loader can later extend them in one place.
func run(path string) error {
	fileTypes := filetype.NewRegistry(xml.Type())
	doc, err := document.Open(path, fileTypes.Detect(path))
	if err != nil {
		return err
	}
	commands := command.NewRegistry(command.Builtins()...)
	// Tab switches between the matchers in this order, the first is the default.
	matchers := []search.Matcher{search.Text{}, search.Fuzzy{}}
	model := app.New(doc, commands, keymap.Default(), matchers, config.Default(), theme.Default())
	_, err = tea.NewProgram(model).Run()
	return err
}
