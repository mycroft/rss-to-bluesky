package db

import (
	"os"
	"path/filepath"
	"testing"
)

// The database stores session tokens, so it must not be readable or writable
// by anyone but its owner.
func TestOpenCreatesPrivateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")

	db, err := openAt(path)
	if err != nil {
		t.Fatalf("openAt: %v", err)
	}
	defer db.Close()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	// A stricter umask may clear more bits, which is fine; group and other
	// must simply have none.
	if perm := info.Mode().Perm(); perm&0077 != 0 {
		t.Errorf("database created with mode %04o, want no group or other access", perm)
	}
}
