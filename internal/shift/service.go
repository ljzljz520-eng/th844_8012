package shift

import (
	"errors"
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

func (s *Service) Start(id, leader string) (model.ShiftSummary, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(leader) == "" {
		return model.ShiftSummary{}, errors.New("shift id and leader are required")
	}
	if _, err := s.store.GetShift(id); err == nil {
		return model.ShiftSummary{}, errors.New("shift already exists")
	}
	shift := model.ShiftSummary{ShiftID: id, Leader: leader, StartedAt: s.now()}
	if err := s.store.PutShift(shift); err != nil {
		return model.ShiftSummary{}, err
	}
	return shift, nil
}

func (s *Service) End(id, notes string) (model.ShiftSummary, error) {
	shift, err := s.store.GetShift(id)
	if err != nil {
		return model.ShiftSummary{}, err
	}
	if shift.EndedAt != nil {
		return shift, errors.New("shift already ended")
	}
	ended := s.now()
	shift.EndedAt = &ended
	shift.Notes = notes
	if err := s.store.PutShift(shift); err != nil {
		return model.ShiftSummary{}, err
	}
	return shift, nil
}

func (s *Service) Current(id string) (model.ShiftSummary, error) { return s.store.GetShift(id) }
