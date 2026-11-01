package app

import (
	"os"
	"path/filepath"
	"strings"
)

// completePath completes path like a shell does: to the longest start that
// the names in its directory share, of those that start like its last
// part. Only directories, which get a "/", and XSD files count. resolve
// turns a path as typed into one to read.
func completePath(path string, resolve func(string) string) string {
	cut := strings.LastIndexAny(path, "/"+string(filepath.Separator)) + 1
	dir, prefix := path[:cut], path[cut:]
	lookIn := resolve(dir)
	if dir == "" {
		lookIn = resolve(".")
	}
	entries, err := os.ReadDir(lookIn)
	if err != nil {
		return path
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}
		if info, err := os.Stat(filepath.Join(lookIn, name)); err == nil && info.IsDir() {
			names = append(names, name+"/")
		} else if strings.EqualFold(filepath.Ext(name), ".xsd") {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return path
	}
	return dir + commonStart(names)
}

// commonStart returns the longest start all names share.
func commonStart(names []string) string {
	common := []rune(names[0])
	for _, name := range names[1:] {
		runes := []rune(name)
		n := 0
		for n < len(common) && n < len(runes) && common[n] == runes[n] {
			n++
		}
		common = common[:n]
	}
	return string(common)
}
