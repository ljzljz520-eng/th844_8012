package grading

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/catalog"
	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

func TestRecordGradeMovesBatchToGrading(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	catalogService := catalog.New(store, func() time.Time { return time.Unix(1, 0) })
	if _, err := catalogService.Receive(model.InboundCommand{ID: "b1", Crop: "橙", ReceivedKg: 100}); err != nil {
		t.Fatal(err)
	}
	service := New(store, catalogService, func() time.Time { return time.Unix(2, 0) })
	if _, err := service.Record(model.GradeCommand{BatchID: "b1", Grade: "one", WeightKg: 40, Inspector: "i"}); err != nil {
		t.Fatal(err)
	}
	batch, err := catalogService.Get("b1")
	if err != nil || batch.Status != "grading" {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
	totals := Totals([]model.GradeRecord{{Grade: "one", WeightKg: 4}, {Grade: "loss", WeightKg: 2}})
	if totals["one"] != 4 || totals["loss"] != 2 {
		t.Fatalf("totals=%v", totals)
	}
}
