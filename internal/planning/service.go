package planning

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"agri-packaging/internal/model"
	"agri-packaging/internal/storage"
)

type LineCapacity struct {
	LineID       string `json:"line_id"`
	BoxesPerHour int    `json:"boxes_per_hour"`
	Priority     int    `json:"priority"`
}

type PlanRequest struct {
	Crop           string `json:"crop"`
	PendingKg      int    `json:"pending_kg"`
	BoxWeightKg    int    `json:"box_weight_kg"`
	HoursAvailable int    `json:"hours_available"`
}

type Plan struct {
	Crop            string  `json:"crop"`
	BoxesNeeded     int     `json:"boxes_needed"`
	RecommendedLine string  `json:"recommended_line"`
	HoursRequired   float64 `json:"hours_required"`
	Feasible        bool    `json:"feasible"`
	Reason          string  `json:"reason"`
}

type Service struct {
	store      *storage.Store
	capacities map[string]LineCapacity
}

func New(store *storage.Store) *Service {
	return &Service{store: store, capacities: make(map[string]LineCapacity)}
}

func (s *Service) SetCapacity(capacity LineCapacity) error {
	if strings.TrimSpace(capacity.LineID) == "" {
		return model.ErrInvalidID
	}
	if capacity.BoxesPerHour <= 0 {
		return errors.New("capacity must be positive")
	}
	if capacity.Priority < 0 {
		return errors.New("priority cannot be negative")
	}
	s.capacities[capacity.LineID] = capacity
	return nil
}

func (s *Service) RemoveCapacity(lineID string) { delete(s.capacities, lineID) }

func (s *Service) Capacities() []LineCapacity {
	result := make([]LineCapacity, 0, len(s.capacities))
	for _, capacity := range s.capacities {
		result = append(result, capacity)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority == result[j].Priority {
			return result[i].LineID < result[j].LineID
		}
		return result[i].Priority < result[j].Priority
	})
	return result
}

func (s *Service) Build(request PlanRequest) (Plan, error) {
	if strings.TrimSpace(request.Crop) == "" {
		return Plan{}, errors.New("crop is required")
	}
	if request.PendingKg <= 0 || request.BoxWeightKg <= 0 {
		return Plan{}, errors.New("weights must be positive")
	}
	if request.HoursAvailable <= 0 {
		return Plan{}, errors.New("hours available must be positive")
	}
	boxes := (request.PendingKg + request.BoxWeightKg - 1) / request.BoxWeightKg
	plan := Plan{Crop: request.Crop, BoxesNeeded: boxes}
	lineIDs, err := s.runningLineIDs()
	if err != nil {
		return Plan{}, err
	}
	capacity, ok := s.chooseCapacity(lineIDs)
	if !ok {
		plan.Reason = "没有可用的运行设备"
		return plan, nil
	}
	plan.RecommendedLine = capacity.LineID
	plan.HoursRequired = float64(boxes) / float64(capacity.BoxesPerHour)
	plan.Feasible = plan.HoursRequired <= float64(request.HoursAvailable)
	if !plan.Feasible {
		plan.Reason = fmt.Sprintf("需要 %.1f 小时，超过可用时长", plan.HoursRequired)
	}
	return plan, nil
}

func (s *Service) runningLineIDs() ([]string, error) {
	lines, err := s.store.ListLines()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(lines))
	for _, item := range lines {
		if !item.Stopped {
			result = append(result, item.ID)
		}
	}
	return result, nil
}

func (s *Service) chooseCapacity(ids []string) (LineCapacity, bool) {
	var best LineCapacity
	found := false
	for _, id := range ids {
		capacity, ok := s.capacities[id]
		if !ok {
			continue
		}
		if !found || capacity.Priority < best.Priority || capacity.Priority == best.Priority && capacity.BoxesPerHour > best.BoxesPerHour {
			best, found = capacity, true
		}
	}
	return best, found
}

func (s *Service) Explain(plan Plan) string {
	if plan.RecommendedLine == "" {
		return "没有可用设备，无法安排"
	}
	if plan.Feasible {
		return fmt.Sprintf("安排 %s，预计 %.1f 小时完成 %d 箱", plan.RecommendedLine, plan.HoursRequired, plan.BoxesNeeded)
	}
	return fmt.Sprintf("安排 %s 但超时：%.1f 小时", plan.RecommendedLine, plan.HoursRequired)
}
