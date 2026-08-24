package planning

import (
	"errors"
	"math"
	"sort"
)

type Allocation struct {
	LineID string  `json:"line_id"`
	Boxes  int     `json:"boxes"`
	Hours  float64 `json:"hours"`
}

func Allocate(boxes int, capacities []LineCapacity) ([]Allocation, error) {
	if boxes <= 0 {
		return nil, errors.New("boxes must be positive")
	}
	usable := make([]LineCapacity, 0, len(capacities))
	for _, capacity := range capacities {
		if capacity.BoxesPerHour > 0 && capacity.LineID != "" {
			usable = append(usable, capacity)
		}
	}
	if len(usable) == 0 {
		return nil, errors.New("no capacity")
	}
	sort.Slice(usable, func(i, j int) bool { return usable[i].BoxesPerHour > usable[j].BoxesPerHour })
	result := make([]Allocation, 0, len(usable))
	remaining := boxes
	for _, capacity := range usable {
		if remaining <= 0 {
			break
		}
		allocation := capacity.BoxesPerHour
		if allocation > remaining {
			allocation = remaining
		}
		result = append(result, Allocation{LineID: capacity.LineID, Boxes: allocation, Hours: float64(allocation) / float64(capacity.BoxesPerHour)})
		remaining -= allocation
	}
	if remaining > 0 {
		return result, errors.New("capacity is insufficient")
	}
	return result, nil
}

func RoundHours(value float64) float64 { return math.Ceil(value*10) / 10 }

func TotalHours(items []Allocation) float64 {
	var total float64
	for _, item := range items {
		total += item.Hours
	}
	return RoundHours(total)
}

func IsOvertime(hours float64, available int) bool { return hours > float64(available) }
