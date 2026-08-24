package report

import (
	"sort"
	"strings"
	"time"

	"agri-packaging/internal/model"
)

type GradeMix struct {
	OnePercent  float64 `json:"one_percent"`
	TwoPercent  float64 `json:"two_percent"`
	LossPercent float64 `json:"loss_percent"`
}

type AuditTimeline struct {
	Hour    string `json:"hour"`
	Actions int    `json:"actions"`
}

func RankRows(rows []CropRow, metric string) []CropRow {
	result := append([]CropRow(nil), rows...)
	sort.SliceStable(result, func(i, j int) bool {
		switch strings.ToLower(metric) {
		case "pending":
			return result[i].PendingKg > result[j].PendingKg
		case "loss":
			return result[i].LossKg > result[j].LossKg
		case "packed":
			return result[i].PackedBoxes > result[j].PackedBoxes
		case "yield":
			return result[i].YieldPercent > result[j].YieldPercent
		default:
			return result[i].Crop < result[j].Crop
		}
	})
	return result
}

func Mix(row CropRow) GradeMix {
	if row.ReceivedKg <= 0 {
		return GradeMix{}
	}
	base := float64(row.ReceivedKg)
	return GradeMix{OnePercent: float64(row.GradeOneKg) / base * 100, TwoPercent: float64(row.GradeTwoKg) / base * 100, LossPercent: float64(row.LossKg) / base * 100}
}

func Timeline(events []model.AuditEvent) []AuditTimeline {
	counts := make(map[string]int)
	for _, event := range events {
		counts[event.CreatedAt.Format("15:00")]++
	}
	result := make([]AuditTimeline, 0, len(counts))
	for hour, count := range counts {
		result = append(result, AuditTimeline{Hour: hour, Actions: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Hour < result[j].Hour })
	return result
}

func Between(events []model.AuditEvent, start, end time.Time) []model.AuditEvent {
	result := make([]model.AuditEvent, 0)
	for _, event := range events {
		if !event.CreatedAt.Before(start) && event.CreatedAt.Before(end) {
			result = append(result, event)
		}
	}
	return result
}

func CropFilter(rows []CropRow, crop string) []CropRow {
	if strings.TrimSpace(crop) == "" {
		return append([]CropRow(nil), rows...)
	}
	result := make([]CropRow, 0)
	for _, row := range rows {
		if row.Crop == crop {
			result = append(result, row)
		}
	}
	return result
}

func Normalize(report DailyReport) DailyReport {
	copyReport := report
	copyReport.Rows = append([]CropRow(nil), report.Rows...)
	for index := range copyReport.Rows {
		row := &copyReport.Rows[index]
		if row.ReceivedKg < 0 {
			row.ReceivedKg = 0
		}
		if row.GradeOneKg < 0 {
			row.GradeOneKg = 0
		}
		if row.GradeTwoKg < 0 {
			row.GradeTwoKg = 0
		}
		if row.LossKg < 0 {
			row.LossKg = 0
		}
		row.PendingKg = row.ReceivedKg - row.GradeOneKg - row.GradeTwoKg - row.LossKg
		if row.PendingKg < 0 {
			row.PendingKg = 0
		}
	}
	return copyReport
}
