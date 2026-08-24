package shift

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/storage"
)

func TestShiftStartAndEnd(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	n := time.Unix(11, 0)
	service := New(store, func() time.Time { return n })
	if _, err := service.Start("shift-1", "leader"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.End("shift-1", "done"); err != nil {
		t.Fatal(err)
	}
	current, err := service.Current("shift-1")
	if err != nil || current.EndedAt == nil || current.Notes != "done" {
		t.Fatalf("current=%+v err=%v", current, err)
	}
}
