package catalog

import (
	"sort"

	"agri-packaging/internal/model"
)

type CropMetric struct {
	Crop       string `json:"crop"`
	ReceivedKg int    `json:"received_kg"`
	GradedKg   int    `json:"graded_kg"`
	PendingKg  int    `json:"pending_kg"`
}

func SummarizeBatches(batches []model.ProduceBatch, grades []model.GradeRecord) []CropMetric {
	metrics := make(map[string]CropMetric)
	for _, batch := range batches {
		metric := metrics[batch.Crop]
		metric.Crop = batch.Crop
		metric.ReceivedKg += batch.ReceivedKg
		metrics[batch.Crop] = metric
	}
	for _, grade := range grades {
		metric := metrics[grade.Crop]
		metric.Crop = grade.Crop
		metric.GradedKg += grade.WeightKg
		metrics[grade.Crop] = metric
	}
	result := make([]CropMetric, 0, len(metrics))
	for _, metric := range metrics {
		metric.PendingKg = metric.ReceivedKg - metric.GradedKg
		if metric.PendingKg < 0 {
			metric.PendingKg = 0
		}
		result = append(result, metric)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Crop < result[j].Crop })
	return result
}
