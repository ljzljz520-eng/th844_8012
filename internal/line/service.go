package line

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

var ErrLineStopped = errors.New("packing line is stopped")
var ErrLineMissing = errors.New("packing line is missing")

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

func (s *Service) Register(id, name string) (model.DeviceLine, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return model.DeviceLine{}, model.ErrInvalidID
	}
	if _, err := s.store.GetLine(id); err == nil {
		return model.DeviceLine{}, errors.New("line already exists")
	}
	line := model.DeviceLine{ID: id, Name: name, UpdatedAt: s.now()}
	if err := s.store.PutLine(line); err != nil {
		return model.DeviceLine{}, fmt.Errorf("save line: %w", err)
	}
	return line, nil
}

func (s *Service) Get(id string) (model.DeviceLine, error) {
	line, err := s.store.GetLine(id)
	if err != nil {
		return model.DeviceLine{}, ErrLineMissing
	}
	return line, nil
}

func (s *Service) Stop(id string, command model.StopCommand) (model.DeviceLine, error) {
	if err := command.Validate(); err != nil {
		return model.DeviceLine{}, err
	}
	line, err := s.Get(id)
	if err != nil {
		return model.DeviceLine{}, err
	}
	if line.Stopped && line.Reason == command.Reason {
		return line, nil
	}
	line.Stopped = true
	line.Reason = command.Reason
	line.UpdatedAt = s.now()
	if err := s.store.PutLine(line); err != nil {
		return model.DeviceLine{}, err
	}
	return line, nil
}

func (s *Service) Resume(id, actor string) (model.DeviceLine, error) {
	if strings.TrimSpace(actor) == "" {
		return model.DeviceLine{}, errors.New("actor is required")
	}
	line, err := s.Get(id)
	if err != nil {
		return model.DeviceLine{}, err
	}
	if !line.Stopped {
		return line, nil
	}
	line.Stopped = false
	line.Reason = ""
	line.UpdatedAt = s.now()
	if err := s.store.PutLine(line); err != nil {
		return model.DeviceLine{}, err
	}
	return line, nil
}

func (s *Service) List() ([]model.DeviceLine, error) { return s.store.ListLines() }

func (s *Service) RequireRunning(id string) error {
	line, err := s.Get(id)
	if err != nil {
		return err
	}
	if line.Stopped {
		return ErrLineStopped
	}
	return nil
}
