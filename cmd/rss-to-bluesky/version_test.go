package main

import (
	"strings"
	"testing"
)

func TestBuildVersion(t *testing.T) {
	originalVersion, originalCommit := version, commit
	t.Cleanup(func() { version, commit = originalVersion, originalCommit })

	t.Run("unstamped builds report devel", func(t *testing.T) {
		version, commit = "", ""
		got := buildVersion()
		if !strings.HasPrefix(got, "rss-to-bluesky devel ") {
			t.Errorf("got %q, want a devel version", got)
		}
	})

	t.Run("stamped version is used", func(t *testing.T) {
		version, commit = "v1.2.3", "0123456789abcdef0123456789abcdef01234567"
		got := buildVersion()
		if !strings.Contains(got, "v1.2.3") {
			t.Errorf("got %q, want it to contain the stamped version", got)
		}
	})

	t.Run("commit is truncated", func(t *testing.T) {
		version, commit = "v1.2.3", "0123456789abcdef0123456789abcdef01234567"
		got := buildVersion()
		if !strings.Contains(got, "0123456789ab") {
			t.Errorf("got %q, want the short commit", got)
		}
		if strings.Contains(got, "0123456789abc") {
			t.Errorf("got %q, commit was not truncated to 12 chars", got)
		}
	})
}
