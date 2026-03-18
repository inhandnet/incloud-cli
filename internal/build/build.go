package build

import (
	"runtime"
)

// These variables are set via ldflags at build time.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func GoVersion() string {
	return runtime.Version()
}
