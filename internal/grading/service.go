package grading

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"agri-packaging/internal/catalog"
	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

type Clock func() time.Time

type Service struct {
	store   *storage.Store
	catalog *catalog.Service
	now     Clock
}

func New(store *storage.Store, catalogService *catalog.Service, now Clock) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, catalog: catalogService, now: now}
}

func (s *Service) Record(cmd model.GradeCommand) (model.GradeRecord, error) {
	if err := cmd.Validate(); err != nil {
		return model.GradeRecord{}, err
	}
	batch, err := s.catalog.Get(cmd.BatchID)
	if err != nil {
		return model.GradeRecord{}, fmt.Errorf("find batch: %w", err)
	}
	if batch.Status == "closed" {
		return model.GradeRecord{}, errors.New("closed batch cannot receive grade")
	}
	if cmd.WeightKg > batch.ReceivedKg {
		return model.GradeRecord{}, errors.New("grade exceeds received weight")
	}
	record := model.GradeRecord{ID: fmt.Sprintf("%s-%s-%d", cmd.BatchID, cmd.Grade, s.now().UnixNano()), BatchID: cmd.BatchID, Crop: batch.Crop, Grade: cmd.Grade, WeightKg: cmd.WeightKg, Inspector: strings.TrimSpace(cmd.Inspector), CreatedAt: s.now()}
	if err := s.store.PutGrade(record); err != nil {
		return model.GradeRecord{}, fmt.Errorf("save grade: %w", err)
	}
	if batch.Status == "received" {
		if err := s.catalog.MarkGrading(batch.ID); err != nil {
			return model.GradeRecord{}, err
		}
	}
	return record, nil
}

func (s *Service) List() ([]model.GradeRecord, error) { return s.store.ListGrades() }

func (s *Service) ByBatch(batchID string) ([]model.GradeRecord, error) {
	if strings.TrimSpace(batchID) == "" {
		return nil, model.ErrInvalidID
	}
	items, err := s.List()
	if err != nil {
		return nil, err
	}
	result := make([]model.GradeRecord, 0)
	for _, item := range items {
		if item.BatchID == batchID {
			result = append(result, item)
		}
	}
	return result, nil
}

func Totals(records []model.GradeRecord) map[string]int {
	totals := map[string]int{"one": 0, "two": 0, "loss": 0}
	for _, record := range records {
		if _, ok := totals[record.Grade]; ok {
			totals[record.Grade] += record.WeightKg
		}
	}
	return totals
}
