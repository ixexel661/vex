package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const header = "# vex settings. Press s in vex to change them.\n\n"

// Dir is where vex keeps its settings, e.g. ~/.config/vex on Linux.
func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vex"), nil
}

// Load reads the config file at path, a missing file gives the defaults. The
// config is usable even with an error: if the file cannot be read, all
// settings keep their defaults, otherwise only the ones the error names.
func Load(path string) (Config, error) {
	c := Default()
	md, err := toml.DecodeFile(path, &c)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), inFile(path, err)
	}
	var errs []error
	for _, key := range md.Undecoded() {
		errs = append(errs, fmt.Errorf("unknown setting %s", key))
	}
	if err := errors.Join(append(errs, c.validate())...); err != nil {
		return c, inFile(path, err)
	}
	return c, nil
}

// Save writes c to the config file at path and creates its directory.
func Save(path string, c Config) error {
	var b bytes.Buffer
	b.WriteString(header)
	enc := toml.NewEncoder(&b)
	enc.Indent = ""
	if err := enc.Encode(c); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// Install writes the default config file if there is none yet.
func Install(path string) error {
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return Save(path, Default())
}

func inFile(path string, err error) error {
	return fmt.Errorf("%s: %s", filepath.Base(path), strings.TrimPrefix(err.Error(), "toml: "))
}
