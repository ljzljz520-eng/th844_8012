package report

import (
	"errors"
	"math"
	"sort"
	"strings"

	"agri-packaging/internal/model"
)

type QualityBand struct {
	Name     string  `json:"name"`
	WeightKg int     `json:"weight_kg"`
	Percent  float64 `json:"percent"`
	Color    string  `json:"color"`
}

type QualitySummary struct {
	Crop     string        `json:"crop"`
	TotalKg  int           `json:"total_kg"`
	Bands    []QualityBand `json:"bands"`
	Balanced bool          `json:"balanced"`
	Message  string        `json:"message"`
}

func BuildQualitySummary(crop string, records []model.GradeRecord) (QualitySummary, error) {
	if strings.TrimSpace(crop) == "" {
		return QualitySummary{}, errors.New("crop is required")
	}
	weights := map[string]int{"one": 0, "two": 0, "loss": 0}
	for _, record := range records {
		if record.Crop != crop {
			continue
		}
		if record.WeightKg <= 0 {
			continue
		}
		if _, ok := weights[record.Grade]; ok {
			weights[record.Grade] += record.WeightKg
		}
	}
	total := weights["one"] + weights["two"] + weights["loss"]
	result := QualitySummary{Crop: crop, TotalKg: total, Bands: make([]QualityBand, 0, 3), Balanced: true}
	for _, band := range []struct {
		name  string
		color string
	}{{"one", "#2f9e6f"}, {"two", "#e6a23c"}, {"loss", "#d35c4e"}} {
		percent := 0.0
		if total > 0 {
			percent = float64(weights[band.name]) / float64(total) * 100
		}
		result.Bands = append(result.Bands, QualityBand{Name: band.name, WeightKg: weights[band.name], Percent: percent, Color: band.color})
	}
	if total == 0 {
		result.Balanced = false
		result.Message = "暂无分级数据"
		return result, nil
	}
	if weights["loss"] > total/3 {
		result.Balanced = false
		result.Message = "损耗比例超过三分之一"
	} else if weights["one"] < total/5 {
		result.Balanced = false
		result.Message = "一级品比例偏低"
	} else {
		result.Message = "品质结构稳定"
	}
	return result, nil
}

func QualityScore(summary QualitySummary) int {
	if summary.TotalKg == 0 {
		return 0
	}
	score := 0.0
	for _, band := range summary.Bands {
		switch band.Name {
		case "one":
			score += band.Percent * 1.2
		case "two":
			score += band.Percent * 0.7
		case "loss":
			score -= band.Percent * 0.8
		}
	}
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return int(math.Round(score))
}

func SortQualitySummaries(items []QualitySummary) []QualitySummary {
	result := append([]QualitySummary(nil), items...)
	sort.SliceStable(result, func(i, j int) bool { return QualityScore(result[i]) > QualityScore(result[j]) })
	return result
}

func CompareQuality(before, after QualitySummary) string {
	change := QualityScore(after) - QualityScore(before)
	if change > 0 {
		return "品质评分提升"
	}
	if change < 0 {
		return "品质评分下降"
	}
	return "品质评分持平"
}

func Band(summary QualitySummary, name string) (QualityBand, bool) {
	for _, band := range summary.Bands {
		if band.Name == name {
			return band, true
		}
	}
	return QualityBand{}, false
}

func LossRatio(summary QualitySummary) float64 {
	band, ok := Band(summary, "loss")
	if !ok {
		return 0
	}
	return band.Percent / 100
}

func MeetsTarget(summary QualitySummary, minimumScore int) bool {
	return QualityScore(summary) >= minimumScore
}

func MergeQualitySummaries(items []QualitySummary) QualitySummary {
	if len(items) == 0 {
		return QualitySummary{}
	}
	result := QualitySummary{Crop: "合计", Bands: make([]QualityBand, 0, 3)}
	weights := map[string]int{"one": 0, "two": 0, "loss": 0}
	for _, item := range items {
		result.TotalKg += item.TotalKg
		for _, band := range item.Bands {
			weights[band.Name] += band.WeightKg
		}
	}
	for _, name := range []string{"one", "two", "loss"} {
		percent := 0.0
		if result.TotalKg > 0 {
			percent = float64(weights[name]) / float64(result.TotalKg) * 100
		}
		result.Bands = append(result.Bands, QualityBand{Name: name, WeightKg: weights[name], Percent: percent})
	}
	result.Balanced = QualityScore(result) >= 60
	return result
}

func QualityLabels(summary QualitySummary) []string {
	labels := make([]string, 0, len(summary.Bands))
	for _, band := range summary.Bands {
		label := band.Name
		if band.Percent >= 50 {
			label += " 主导"
		} else if band.Percent == 0 {
			label += " 无记录"
		}
		labels = append(labels, label)
	}
	return labels
}

func ClampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func BandNames(summary QualitySummary) []string {
	result := make([]string, 0, len(summary.Bands))
	for _, band := range summary.Bands {
		result = append(result, band.Name)
	}
	return result
}

func EmptyQuality(crop string) QualitySummary {
	return QualitySummary{Crop: crop, Bands: []QualityBand{}, Balanced: false, Message: "暂无分级数据"}
}

func TotalBandWeight(summary QualitySummary) int {
	total := 0
	for _, band := range summary.Bands {
		total += band.WeightKg
	}
	return total
}

func SummaryReady(summary QualitySummary) bool {
	return summary.TotalKg > 0 && TotalBandWeight(summary) > 0
}

func ScoreLabel(score int) string {
	if score >= 85 {
		return "优秀"
	}
	if score >= 60 {
		return "合格"
	}
	return "需关注"
}
