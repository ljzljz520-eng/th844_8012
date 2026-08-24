package storage

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "screen.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	want := model.ProduceBatch{ID: "batch-1", Crop: "苹果", Variety: "富士", ReceivedKg: 1200, Status: "received", CreatedAt: time.Unix(10, 0)}
	if err := store.PutBatch(want); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	got, err := reopened.GetBatch("batch-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.ReceivedKg != want.ReceivedKg || got.Crop != want.Crop {
		t.Fatalf("got %+v", got)
	}
}

func TestStoreListsAndDeletesProgress(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "screen.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	item := model.PackingProgress{ID: "batch-1/line-a", BatchID: "batch-1", LineID: "line-a", Boxes: 4, TargetBoxes: 10}
	if err := store.PutProgress(item); err != nil {
		t.Fatal(err)
	}
	items, err := store.ListProgress()
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if err := store.DeleteProgress(item.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetProgress(item.ID); err == nil {
		t.Fatal("expected deleted progress")
	}
}
