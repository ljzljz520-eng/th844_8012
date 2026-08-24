package planning

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Alert struct {
	Code      string    `json:"code"`
	Level     string    `json:"level"`
	LineID    string    `json:"line_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func DetectAlerts(now time.Time, lines []LineCapacity, stopped map[string]bool, schedule Schedule) []Alert {
	result := make([]Alert, 0)
	for _, capacity := range lines {
		if strings.TrimSpace(capacity.LineID) == "" {
			continue
		}
		if stopped[capacity.LineID] {
			result = append(result, Alert{Code: "LINE_STOPPED", Level: "warning", LineID: capacity.LineID, Message: "设备停机，需重新安排装箱", CreatedAt: now})
		}
	}
	for _, entry := range schedule.Entries {
		if entry.End.Before(now) {
			result = append(result, Alert{Code: "OVERDUE", Level: "critical", LineID: entry.LineID, Message: fmt.Sprintf("%s 的 %d 箱任务已超时", entry.Crop, entry.Boxes), CreatedAt: now})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Level == result[j].Level {
			return result[i].LineID < result[j].LineID
		}
		return severity(result[i].Level) > severity(result[j].Level)
	})
	return result
}

func severity(level string) int {
	switch level {
	case "critical":
		return 3
	case "warning":
		return 2
	default:
		return 1
	}
}

func DedupeAlerts(alerts []Alert) []Alert {
	seen := make(map[string]bool)
	result := make([]Alert, 0, len(alerts))
	for _, alert := range alerts {
		key := alert.Code + ":" + alert.LineID
		if !seen[key] {
			seen[key] = true
			result = append(result, alert)
		}
	}
	return result
}

func AlertSummary(alerts []Alert) string {
	if len(alerts) == 0 {
		return "运行正常"
	}
	critical, warning := 0, 0
	for _, alert := range alerts {
		if alert.Level == "critical" {
			critical++
		} else if alert.Level == "warning" {
			warning++
		}
	}
	if critical > 0 {
		return fmt.Sprintf("%d 项严重告警，%d 项提醒", critical, warning)
	}
	return fmt.Sprintf("%d 项提醒", warning)
}

func FilterByLevel(alerts []Alert, level string) []Alert {
	result := make([]Alert, 0)
	for _, alert := range alerts {
		if level == "" || alert.Level == level {
			result = append(result, alert)
		}
	}
	return result
}

func HasCritical(alerts []Alert) bool {
	for _, alert := range alerts {
		if alert.Level == "critical" {
			return true
		}
	}
	return false
}

func ResolveLine(alerts []Alert, lineID string) []Alert {
	result := make([]Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert.LineID != lineID {
			result = append(result, alert)
		}
	}
	return result
}

func SortByTime(alerts []Alert) []Alert {
	result := append([]Alert(nil), alerts...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result
}

func AlertCodes(alerts []Alert) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, alert := range alerts {
		if !seen[alert.Code] {
			seen[alert.Code] = true
			result = append(result, alert.Code)
		}
	}
	sort.Strings(result)
	return result
}
