package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"agri-packaging/internal/model"
	"go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("record not found")

var bucketNames = [][]byte{
	[]byte("ProduceBatch"), []byte("GradeRecord"), []byte("PackingProgress"),
	[]byte("DeviceLine"), []byte("AuditEvent"), []byte("DashboardSnapshot"),
	[]byte("OperatorSetting"), []byte("ShiftSummary"),
}

type Store struct {
	db *bbolt.DB
	mu sync.RWMutex
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage path is required")
	}
	db, err := bbolt.Open(filepath.Clean(path), 0600, &bbolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}
	store := &Store{db: db}
	if err := store.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func encode(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode value: %w", err)
	}
	return data, nil
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return ErrNotFound
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode value: %w", err)
	}
	return nil
}

func put(tx *bbolt.Tx, bucket, key string, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	return b.Put([]byte(key), data)
}

func get(tx *bbolt.Tx, bucket, key string, target any) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	return decode(b.Get([]byte(key)), target)
}

func deleteKey(tx *bbolt.Tx, bucket, key string) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	return b.Delete([]byte(key))
}

func list[T any](s *Store, bucket string) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var values []T
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s missing", bucket)
		}
		return b.ForEach(func(_, value []byte) error {
			var item T
			if err := decode(value, &item); err != nil {
				return err
			}
			values = append(values, item)
			return nil
		})
	})
	return values, err
}

func (s *Store) PutBatch(batch model.ProduceBatch) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "ProduceBatch", batch.ID, batch) })
}

func (s *Store) GetBatch(id string) (model.ProduceBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var batch model.ProduceBatch
	err := s.db.View(func(tx *bbolt.Tx) error { return get(tx, "ProduceBatch", id, &batch) })
	return batch, err
}

func (s *Store) ListBatches() ([]model.ProduceBatch, error) {
	return list[model.ProduceBatch](s, "ProduceBatch")
}

func (s *Store) PutGrade(record model.GradeRecord) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "GradeRecord", record.ID, record) })
}

func (s *Store) ListGrades() ([]model.GradeRecord, error) {
	return list[model.GradeRecord](s, "GradeRecord")
}

func (s *Store) PutProgress(progress model.PackingProgress) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "PackingProgress", progress.ID, progress) })
}

func (s *Store) GetProgress(id string) (model.PackingProgress, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var progress model.PackingProgress
	err := s.db.View(func(tx *bbolt.Tx) error { return get(tx, "PackingProgress", id, &progress) })
	return progress, err
}

func (s *Store) DeleteProgress(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return deleteKey(tx, "PackingProgress", id) })
}

func (s *Store) ListProgress() ([]model.PackingProgress, error) {
	return list[model.PackingProgress](s, "PackingProgress")
}

func (s *Store) PutLine(line model.DeviceLine) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "DeviceLine", line.ID, line) })
}

func (s *Store) GetLine(id string) (model.DeviceLine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var line model.DeviceLine
	err := s.db.View(func(tx *bbolt.Tx) error { return get(tx, "DeviceLine", id, &line) })
	return line, err
}

func (s *Store) ListLines() ([]model.DeviceLine, error) {
	return list[model.DeviceLine](s, "DeviceLine")
}

func (s *Store) PutAudit(event model.AuditEvent) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "AuditEvent", event.ID, event) })
}

func (s *Store) ListAudits() ([]model.AuditEvent, error) {
	return list[model.AuditEvent](s, "AuditEvent")
}

func (s *Store) PutSnapshot(snapshot model.DashboardSnapshot) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := snapshot.Crop
	if key == "" {
		key = "all"
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "DashboardSnapshot", key, snapshot) })
}

func (s *Store) ListSnapshots() ([]model.DashboardSnapshot, error) {
	return list[model.DashboardSnapshot](s, "DashboardSnapshot")
}

func (s *Store) PutSetting(setting model.OperatorSetting) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "OperatorSetting", setting.Key, setting) })
}

func (s *Store) GetSetting(key string) (model.OperatorSetting, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var setting model.OperatorSetting
	err := s.db.View(func(tx *bbolt.Tx) error { return get(tx, "OperatorSetting", key, &setting) })
	return setting, err
}

func (s *Store) PutShift(shift model.ShiftSummary) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return put(tx, "ShiftSummary", shift.ShiftID, shift) })
}

func (s *Store) GetShift(id string) (model.ShiftSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var shift model.ShiftSummary
	err := s.db.View(func(tx *bbolt.Tx) error { return get(tx, "ShiftSummary", id, &shift) })
	return shift, err
}
