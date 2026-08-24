package planning

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

type ScheduleEntry struct {
	LineID string    `json:"line_id"`
	Crop   string    `json:"crop"`
	Boxes  int       `json:"boxes"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Reason string    `json:"reason"`
}

type Schedule struct {
	Date    time.Time       `json:"date"`
	Entries []ScheduleEntry `json:"entries"`
}

func BuildSchedule(start time.Time, crop string, boxes int, capacities []LineCapacity) (Schedule, error) {
	if start.IsZero() {
		return Schedule{}, errors.New("start time is required")
	}
	if crop == "" {
		return Schedule{}, errors.New("crop is required")
	}
	allocations, err := Allocate(boxes, capacities)
	if err != nil {
		return Schedule{}, err
	}
	schedule := Schedule{Date: start, Entries: make([]ScheduleEntry, 0, len(allocations))}
	current := start
	for _, allocation := range allocations {
		end := current.Add(time.Duration(allocation.Hours * float64(time.Hour)))
		schedule.Entries = append(schedule.Entries, ScheduleEntry{LineID: allocation.LineID, Crop: crop, Boxes: allocation.Boxes, Start: current, End: end, Reason: fmt.Sprintf("分配 %d 箱", allocation.Boxes)})
		current = end
	}
	return schedule, nil
}

func ValidateSchedule(schedule Schedule) error {
	if len(schedule.Entries) == 0 {
		return errors.New("schedule is empty")
	}
	entries := append([]ScheduleEntry(nil), schedule.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Start.Before(entries[j].Start) })
	for index, entry := range entries {
		if entry.LineID == "" || entry.Boxes <= 0 {
			return errors.New("schedule entry is incomplete")
		}
		if !entry.End.After(entry.Start) {
			return errors.New("schedule entry has invalid time")
		}
		if index > 0 && entry.Start.Before(entries[index-1].End) {
			return errors.New("schedule entries overlap")
		}
	}
	return nil
}

func Duration(schedule Schedule) time.Duration {
	if len(schedule.Entries) == 0 {
		return 0
	}
	start, end := schedule.Entries[0].Start, schedule.Entries[0].End
	for _, entry := range schedule.Entries[1:] {
		if entry.Start.Before(start) {
			start = entry.Start
		}
		if entry.End.After(end) {
			end = entry.End
		}
	}
	return end.Sub(start)
}

func Utilization(schedule Schedule, available time.Duration) float64 {
	if available <= 0 {
		return 0
	}
	return float64(Duration(schedule)) / float64(available) * 100
}

func MergeSchedules(base, extra Schedule) Schedule {
	merged := Schedule{Date: base.Date, Entries: append([]ScheduleEntry(nil), base.Entries...)}
	if merged.Date.IsZero() {
		merged.Date = extra.Date
	}
	merged.Entries = append(merged.Entries, extra.Entries...)
	sort.Slice(merged.Entries, func(i, j int) bool { return merged.Entries[i].Start.Before(merged.Entries[j].Start) })
	return merged
}

func RescheduleAfter(schedule Schedule, moment time.Time) Schedule {
	copySchedule := Schedule{Date: schedule.Date, Entries: append([]ScheduleEntry(nil), schedule.Entries...)}
	for index := range copySchedule.Entries {
		entry := &copySchedule.Entries[index]
		if entry.Start.Before(moment) {
			duration := entry.End.Sub(entry.Start)
			entry.Start = moment
			entry.End = moment.Add(duration)
			moment = entry.End
		}
	}
	return copySchedule
}
