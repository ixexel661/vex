package main

import (
	"regexp"
	"runtime/debug"
)

// version is the tag a release is built from. The release workflow sets it
// with -ldflags "-X main.version=v1.2.0".
var version string

// pseudoVersion matches the versions Go makes up for builds between tags,
// e.g. v0.0.0-20261015123645-882f7945108b.
var pseudoVersion = regexp.MustCompile(`\d{14}-[0-9a-f]{12}`)

func currentVersion() string {
	module := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		module = info.Main.Version
	}
	return versionOf(version, module)
}

// versionOf returns the tag of a release build, or else the module version,
// which "go install github.com/ixexel661/vex@v1.2.0" sets. Any other build is
// "dev".
func versionOf(tagged, module string) string {
	switch {
	case tagged != "":
		return tagged
	case module == "" || module == "(devel)" || pseudoVersion.MatchString(module):
		return "dev"
	default:
		return module
	}
}
