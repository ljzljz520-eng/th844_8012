package catalog

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

var ErrDuplicateBatch = errors.New("batch already exists")

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

func (s *Service) Receive(cmd model.InboundCommand) (model.ProduceBatch, error) {
	if err := cmd.Validate(); err != nil {
		return model.ProduceBatch{}, err
	}
	if _, err := s.store.GetBatch(cmd.ID); err == nil {
		return model.ProduceBatch{}, ErrDuplicateBatch
	}
	batch := model.ProduceBatch{ID: strings.TrimSpace(cmd.ID), Crop: strings.TrimSpace(cmd.Crop), Variety: strings.TrimSpace(cmd.Variety), ReceivedKg: cmd.ReceivedKg, CreatedAt: s.now(), Status: "received"}
	if batch.Variety == "" {
		batch.Variety = "standard"
	}
	if err := s.store.PutBatch(batch); err != nil {
		return model.ProduceBatch{}, fmt.Errorf("save batch: %w", err)
	}
	return batch, nil
}

func (s *Service) Get(id string) (model.ProduceBatch, error) {
	if strings.TrimSpace(id) == "" {
		return model.ProduceBatch{}, model.ErrInvalidID
	}
	return s.store.GetBatch(id)
}

func (s *Service) List() ([]model.ProduceBatch, error) { return s.store.ListBatches() }

func (s *Service) MarkGrading(id string) error {
	batch, err := s.Get(id)
	if err != nil {
		return err
	}
	if batch.Status == "closed" {
		return errors.New("closed batch cannot be graded")
	}
	batch.Status = "grading"
	return s.store.PutBatch(batch)
}

func (s *Service) Close(id string) error {
	batch, err := s.Get(id)
	if err != nil {
		return err
	}
	if batch.Status == "received" {
		return errors.New("batch must be graded before closing")
	}
	batch.Status = "closed"
	return s.store.PutBatch(batch)
}
