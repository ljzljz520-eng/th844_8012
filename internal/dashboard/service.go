package dashboard

import (
	"sort"
	"strings"
	"time"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

type Clock func() time.Time

type Service struct {
	store *storage.Store
	now   Clock
}

func New(store *storage.Store, now Clock) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func (s *Service) Snapshot(crop string) (model.DashboardSnapshot, error) {
	batches, err := s.store.ListBatches()
	if err != nil {
		return model.DashboardSnapshot{}, err
	}
	grades, err := s.store.ListGrades()
	if err != nil {
		return model.DashboardSnapshot{}, err
	}
	progress, err := s.store.ListProgress()
	if err != nil {
		return model.DashboardSnapshot{}, err
	}
	lines, err := s.store.ListLines()
	if err != nil {
		return model.DashboardSnapshot{}, err
	}
	snapshot := model.DashboardSnapshot{Crop: crop, UpdatedAt: s.now()}
	for _, batch := range batches {
		if crop != "" && batch.Crop != crop {
			continue
		}
		snapshot.PendingKg += batch.ReceivedKg
	}
	for _, grade := range grades {
		if crop != "" && grade.Crop != crop {
			continue
		}
		snapshot.PendingKg -= grade.WeightKg
		switch grade.Grade {
		case "one":
			snapshot.GradeOneKg += grade.WeightKg
		case "two":
			snapshot.GradeTwoKg += grade.WeightKg
		case "loss":
			snapshot.LossKg += grade.WeightKg
		}
	}
	if snapshot.PendingKg < 0 {
		snapshot.PendingKg = 0
	}
	for _, item := range progress {
		snapshot.PackedBoxes += item.Boxes
		snapshot.TargetBoxes += item.TargetBoxes
	}
	for _, item := range lines {
		if item.Stopped {
			snapshot.StoppedLines++
		}
	}
	if err := s.store.PutSnapshot(snapshot); err != nil {
		return model.DashboardSnapshot{}, err
	}
	return snapshot, nil
}

func (s *Service) RecentAudits(limit int) ([]model.AuditEvent, error) {
	items, err := s.store.ListAudits()
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if limit <= 0 {
		limit = 10
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *Service) FilterSnapshots(crop string) ([]model.DashboardSnapshot, error) {
	items, err := s.store.ListSnapshots()
	if err != nil {
		return nil, err
	}
	result := make([]model.DashboardSnapshot, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(crop) == "" || item.Crop == crop {
			result = append(result, item)
		}
	}
	return result, nil
}
