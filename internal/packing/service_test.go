package packing

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/line"
	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

func TestPackingRejectsStoppedLine(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	lineService := line.New(store, func() time.Time { return time.Unix(3, 0) })
	if _, err := lineService.Register("line-a", "一号线"); err != nil {
		t.Fatal(err)
	}
	if _, err := lineService.Stop("line-a", model.StopCommand{Reason: "维护", Actor: "leader"}); err != nil {
		t.Fatal(err)
	}
	service := New(store, lineService, func() time.Time { return time.Unix(4, 0) })
	if _, err := service.Submit(model.ProgressCommand{BatchID: "batch-a", LineID: "line-a", Boxes: 5, TargetBoxes: 20}); err == nil {
		t.Fatal("stopped line accepted progress")
	}
	item, err := service.Get("batch-a", "line-a")
	if err == nil && item.Boxes != 0 {
		t.Fatalf("stopped line changed progress: %+v", item)
	}
}

func TestPackingCompletionAndReset(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	lineService := line.New(store, nil)
	if _, err := lineService.Register("line-b", "二号线"); err != nil {
		t.Fatal(err)
	}
	service := New(store, lineService, func() time.Time { return time.Unix(5, 0) })
	item, err := service.Submit(model.ProgressCommand{BatchID: "batch-b", LineID: "line-b", Boxes: 4, TargetBoxes: 10})
	if err != nil || item.Boxes != 4 {
		t.Fatalf("item=%+v err=%v", item, err)
	}
	if service.Completion(item) != 40 {
		t.Fatalf("completion=%v", service.Completion(item))
	}
	if err := service.Reset("batch-b", "line-b"); err != nil {
		t.Fatal(err)
	}
}
