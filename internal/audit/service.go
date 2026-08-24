package audit

import (
	"fmt"
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

func (s *Service) Record(action, actor, entity, entityID, detail string) (model.AuditEvent, error) {
	if strings.TrimSpace(action) == "" || strings.TrimSpace(actor) == "" {
		return model.AuditEvent{}, fmt.Errorf("action and actor are required")
	}
	event := model.AuditEvent{ID: fmt.Sprintf("audit-%d-%s-%s", s.now().UnixNano(), action, entityID), Action: action, Actor: actor, Entity: entity, EntityID: entityID, Detail: detail, CreatedAt: s.now()}
	if err := s.store.PutAudit(event); err != nil {
		return model.AuditEvent{}, err
	}
	return event, nil
}

func (s *Service) List() ([]model.AuditEvent, error) { return s.store.ListAudits() }

func (s *Service) ForEntity(entity, entityID string) ([]model.AuditEvent, error) {
	items, err := s.List()
	if err != nil {
		return nil, err
	}
	result := make([]model.AuditEvent, 0)
	for _, item := range items {
		if item.Entity == entity && item.EntityID == entityID {
			result = append(result, item)
		}
	}
	return result, nil
}
