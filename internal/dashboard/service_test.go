package dashboard

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

func TestSnapshotAggregatesScreenMetrics(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.PutBatch(model.ProduceBatch{ID: "b", Crop: "苹果", ReceivedKg: 100}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutGrade(model.GradeRecord{ID: "g1", Crop: "苹果", Grade: "one", WeightKg: 50}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutGrade(model.GradeRecord{ID: "g2", Crop: "苹果", Grade: "loss", WeightKg: 5}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutLine(model.DeviceLine{ID: "l", Stopped: true}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutProgress(model.PackingProgress{ID: "b/l", Boxes: 2, TargetBoxes: 4}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := New(store, func() time.Time { return time.Unix(8, 0) }).Snapshot("苹果")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.PendingKg != 45 || snapshot.GradeOneKg != 50 || snapshot.LossKg != 5 || snapshot.PackedBoxes != 2 || snapshot.StoppedLines != 1 {
		t.Fatalf("snapshot=%+v", snapshot)
	}
}
