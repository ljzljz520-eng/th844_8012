package catalog

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

func TestReceiveAndLifecycle(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store, func() time.Time { return time.Unix(42, 0) })
	batch, err := service.Receive(model.InboundCommand{ID: "b1", Crop: "梨", ReceivedKg: 900})
	if err != nil {
		t.Fatal(err)
	}
	if batch.Variety != "standard" || batch.CreatedAt.Unix() != 42 {
		t.Fatalf("batch=%+v", batch)
	}
	if err := service.MarkGrading("b1"); err != nil {
		t.Fatal(err)
	}
	if err := service.Close("b1"); err != nil {
		t.Fatal(err)
	}
	if err := service.Close("missing"); err == nil {
		t.Fatal("expected missing error")
	}
}

func TestSummarizeBatches(t *testing.T) {
	batches := []model.ProduceBatch{{ID: "1", Crop: "苹果", ReceivedKg: 100}, {ID: "2", Crop: "苹果", ReceivedKg: 50}}
	grades := []model.GradeRecord{{ID: "g", Crop: "苹果", WeightKg: 80}}
	result := SummarizeBatches(batches, grades)
	if len(result) != 1 || result[0].PendingKg != 70 {
		t.Fatalf("result=%+v", result)
	}
}
