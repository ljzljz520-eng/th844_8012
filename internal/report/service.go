package report

import (
	"io"
	"sort"
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

type CropRow struct {
	Crop         string  `json:"crop"`
	ReceivedKg   int     `json:"received_kg"`
	GradeOneKg   int     `json:"grade_one_kg"`
	GradeTwoKg   int     `json:"grade_two_kg"`
	LossKg       int     `json:"loss_kg"`
	PendingKg    int     `json:"pending_kg"`
	PackedBoxes  int     `json:"packed_boxes"`
	TargetBoxes  int     `json:"target_boxes"`
	YieldPercent float64 `json:"yield_percent"`
}

type DailyReport struct {
	GeneratedAt      time.Time          `json:"generated_at"`
	ShiftID          string             `json:"shift_id"`
	Leader           string             `json:"leader"`
	Rows             []CropRow          `json:"rows"`
	TotalReceivedKg  int                `json:"total_received_kg"`
	TotalGradedKg    int                `json:"total_graded_kg"`
	TotalLossKg      int                `json:"total_loss_kg"`
	TotalPackedBoxes int                `json:"total_packed_boxes"`
	StoppedLines     []model.DeviceLine `json:"stopped_lines"`
	AuditCount       int                `json:"audit_count"`
}

func New(store *storage.Store, now Clock) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func (s *Service) Build(shiftID string) (DailyReport, error) {
	batches, err := s.store.ListBatches()
	if err != nil {
		return DailyReport{}, err
	}
	grades, err := s.store.ListGrades()
	if err != nil {
		return DailyReport{}, err
	}
	progress, err := s.store.ListProgress()
	if err != nil {
		return DailyReport{}, err
	}
	lines, err := s.store.ListLines()
	if err != nil {
		return DailyReport{}, err
	}
	audits, err := s.store.ListAudits()
	if err != nil {
		return DailyReport{}, err
	}
	rows := make(map[string]CropRow)
	for _, batch := range batches {
		row := rows[batch.Crop]
		row.Crop = batch.Crop
		row.ReceivedKg += batch.ReceivedKg
		rows[batch.Crop] = row
	}
	for _, grade := range grades {
		row := rows[grade.Crop]
		row.Crop = grade.Crop
		switch grade.Grade {
		case "one":
			row.GradeOneKg += grade.WeightKg
		case "two":
			row.GradeTwoKg += grade.WeightKg
		case "loss":
			row.LossKg += grade.WeightKg
		}
		rows[grade.Crop] = row
	}
	for _, item := range progress {
		batch, lookupErr := findBatch(batches, item.BatchID)
		if lookupErr != nil {
			continue
		}
		row := rows[batch.Crop]
		row.Crop = batch.Crop
		row.PackedBoxes += item.Boxes
		row.TargetBoxes += item.TargetBoxes
		rows[batch.Crop] = row
	}
	result := DailyReport{GeneratedAt: s.now(), ShiftID: shiftID, Rows: make([]CropRow, 0, len(rows)), StoppedLines: make([]model.DeviceLine, 0), AuditCount: len(audits)}
	for _, line := range lines {
		if line.Stopped {
			result.StoppedLines = append(result.StoppedLines, line)
		}
	}
	for _, row := range rows {
		row.PendingKg = row.ReceivedKg - row.GradeOneKg - row.GradeTwoKg - row.LossKg
		if row.PendingKg < 0 {
			row.PendingKg = 0
		}
		graded := row.GradeOneKg + row.GradeTwoKg
		if row.ReceivedKg > 0 {
			row.YieldPercent = float64(graded) / float64(row.ReceivedKg) * 100
		}
		result.TotalReceivedKg += row.ReceivedKg
		result.TotalGradedKg += graded
		result.TotalLossKg += row.LossKg
		result.TotalPackedBoxes += row.PackedBoxes
		result.Rows = append(result.Rows, row)
	}
	sort.Slice(result.Rows, func(i, j int) bool { return result.Rows[i].Crop < result.Rows[j].Crop })
	if shiftID != "" {
		if shift, shiftErr := s.store.GetShift(shiftID); shiftErr == nil {
			result.Leader = shift.Leader
		}
	}
	return result, nil
}

func findBatch(batches []model.ProduceBatch, id string) (model.ProduceBatch, error) {
	for _, batch := range batches {
		if batch.ID == id {
			return batch, nil
		}
	}
	return model.ProduceBatch{}, storage.ErrNotFound
}

func (s *Service) CropNames() ([]string, error) {
	batches, err := s.store.ListBatches()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	for _, batch := range batches {
		name := strings.TrimSpace(batch.Crop)
		if name != "" {
			seen[name] = true
		}
	}
	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

func (s *Service) AtRiskRows(report DailyReport, threshold float64) []CropRow {
	if threshold <= 0 {
		threshold = 70
	}
	result := make([]CropRow, 0)
	for _, row := range report.Rows {
		if row.YieldPercent < threshold || row.PendingKg > row.ReceivedKg/2 {
			result = append(result, row)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].YieldPercent < result[j].YieldPercent })
	return result
}

func (s *Service) ProgressPercent(report DailyReport) float64 {
	totalTarget := 0
	for _, row := range report.Rows {
		totalTarget += row.TargetBoxes
	}
	if totalTarget == 0 {
		return 0
	}
	return float64(report.TotalPackedBoxes) / float64(totalTarget) * 100
}

func (s *Service) Write(w io.Writer, daily DailyReport, format string) error {
	return WriteReport(w, daily, format)
}
