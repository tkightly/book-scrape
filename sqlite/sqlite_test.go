package sqlite

import (
	"os"
	"testing"
)

func TestNewDb(t *testing.T) {
	path := t.TempDir() + "db.db"

	_, err := NewDb(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

	_, err = os.Stat(path)

	if err != nil {
		t.Errorf("error: %s", err)
	}

}
