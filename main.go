package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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

//go:embed themes/*.toml
var builtinThemes embed.FS

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
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	builtin, err := fs.Sub(builtinThemes, "themes")
	if err != nil {
		return err
	}
	configPath := filepath.Join(dir, "config.toml")
	themes := theme.NewStore(filepath.Join(dir, "themes"), builtin)
	// The first start creates the files, so they are easy to find and change.
	installErr := errors.Join(config.Install(configPath), themes.Install())
	cfg, configErr := config.Load(configPath)
	th, themeErr := themes.Load(cfg.Theme)

	model := app.New(doc, app.Setup{
		Commands:   command.NewRegistry(command.Builtins()...),
		Keys:       keymap.Default(),
		Matchers:   []search.Matcher{search.Text{}, search.Fuzzy{}},
		Config:     cfg,
		ConfigPath: configPath,
		Theme:      th,
		Themes:     themes,
		Warning:    errors.Join(installErr, configErr, themeErr),
	})
	_, err = tea.NewProgram(model).Run()
	return err
}
