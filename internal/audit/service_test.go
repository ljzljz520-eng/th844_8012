package audit

import (
	"path/filepath"
	"testing"
	"time"

	"agri-packaging/internal/storage"
)

func TestAuditRecordAndFilter(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := New(store, func() time.Time { return time.Unix(9, 0) })
	if _, err := service.Record("stop", "leader", "DeviceLine", "line-a", "maintenance"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Record("receive", "operator", "ProduceBatch", "batch-a", "dock"); err != nil {
		t.Fatal(err)
	}
	items, err := service.ForEntity("DeviceLine", "line-a")
	if err != nil || len(items) != 1 || items[0].Action != "stop" {
		t.Fatalf("items=%v err=%v", items, err)
	}
}
