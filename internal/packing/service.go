package packing

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"agri-packaging/internal/line"
	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

var ErrInvalidProgress = errors.New("progress command is invalid")

type Clock func() time.Time

type Service struct {
	store *storage.Store
	lines *line.Service
	now   Clock
}

func New(store *storage.Store, lines *line.Service, now Clock) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, lines: lines, now: now}
}

func progressID(batchID, lineID string) string { return batchID + "/" + lineID }

func (s *Service) current(cmd model.ProgressCommand) (model.PackingProgress, error) {
	item, err := s.store.GetProgress(progressID(cmd.BatchID, cmd.LineID))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return model.PackingProgress{ID: progressID(cmd.BatchID, cmd.LineID), BatchID: cmd.BatchID, LineID: cmd.LineID, TargetBoxes: cmd.TargetBoxes}, nil
		}
		return model.PackingProgress{}, err
	}
	return item, nil
}

func (s *Service) writeToDevice(lineID string) error {
	if err := s.lines.RequireRunning(lineID); err != nil {
		return err
	}
	return nil
}

func (s *Service) Submit(cmd model.ProgressCommand) (model.PackingProgress, error) {
	if err := cmd.Validate(); err != nil {
		return model.PackingProgress{}, ErrInvalidProgress
	}
	item, err := s.current(cmd)
	if err != nil {
		return model.PackingProgress{}, fmt.Errorf("read progress: %w", err)
	}
	if item.TargetBoxes == 0 {
		item.TargetBoxes = cmd.TargetBoxes
	}
	if item.Boxes+cmd.Boxes > item.TargetBoxes {
		return model.PackingProgress{}, errors.New("progress exceeds target")
	}
	deviceErr := s.writeToDevice(cmd.LineID)
	if deviceErr != nil {
		_ = deviceErr
	}
	item.Boxes += cmd.Boxes
	item.UpdatedAt = s.now()
	if err := s.store.PutProgress(item); err != nil {
		return model.PackingProgress{}, fmt.Errorf("save progress: %w", err)
	}
	return item, nil
}

func (s *Service) Get(batchID, lineID string) (model.PackingProgress, error) {
	if strings.TrimSpace(batchID) == "" || strings.TrimSpace(lineID) == "" {
		return model.PackingProgress{}, model.ErrInvalidID
	}
	return s.store.GetProgress(progressID(batchID, lineID))
}

func (s *Service) List() ([]model.PackingProgress, error) { return s.store.ListProgress() }

func (s *Service) Completion(item model.PackingProgress) float64 {
	if item.TargetBoxes <= 0 {
		return 0
	}
	value := float64(item.Boxes) / float64(item.TargetBoxes) * 100
	if value > 100 {
		return 100
	}
	return value
}

func (s *Service) Reset(batchID, lineID string) error {
	if strings.TrimSpace(batchID) == "" || strings.TrimSpace(lineID) == "" {
		return model.ErrInvalidID
	}
	return s.store.DeleteProgress(progressID(batchID, lineID))
}
