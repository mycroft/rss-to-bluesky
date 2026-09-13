package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// version and commit are stamped at build time with -ldflags "-X main.version=...".
// Release builds set version to the git tag. When they are empty the values are
// recovered from the VCS metadata the Go toolchain embeds in the binary, so a
// plain "go build" still reports something useful.
var (
	version string
	commit  string
)

// buildVersion renders the line printed by -version.
func buildVersion() string {
	name := version
	if name == "" {
		name = "devel"
	}

	revision := commit
	// vcs.time is the timestamp of the revision, not of the build.
	revisionTime := "unknown"
	modified := false

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if revision == "" {
					revision = setting.Value
				}
			case "vcs.time":
				revisionTime = setting.Value
			case "vcs.modified":
				modified = setting.Value == "true"
			}
		}
	}

	if revision == "" {
		revision = "unknown"
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if modified {
		revision += "-dirty"
	}

	return fmt.Sprintf("rss-to-bluesky %s (commit %s, %s, %s)",
		name, revision, revisionTime, runtime.Version())
}
