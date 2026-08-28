package line

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

func TestStopAndResumeLine(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store, func() time.Time { return time.Unix(7, 0) })
	if _, err := service.Register("line-a", "一号线"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Stop("line-a", model.StopCommand{Reason: "清理", Actor: "leader"}); err != nil {
		t.Fatal(err)
	}
	if err := service.RequireRunning("line-a"); err != ErrLineStopped {
		t.Fatalf("err=%v", err)
	}
	if _, err := service.Resume("line-a", "leader"); err != nil {
		t.Fatal(err)
	}
	if err := service.RequireRunning("line-a"); err != nil {
		t.Fatal(err)
	}
}

func TestMissingLine(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := New(store, nil).Get("none"); err != ErrLineMissing {
		t.Fatalf("err=%v", err)
	}
}
