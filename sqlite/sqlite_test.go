package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDb(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.db")

	_, err := NewDb(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	_, err = os.Stat(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

}
