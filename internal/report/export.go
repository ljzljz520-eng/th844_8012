package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"agri-packaging/internal/model"
)

func ToCSV(report DailyReport) (string, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"品类", "收货kg", "一级品kg", "二级品kg", "损耗kg", "待分级kg", "装箱数", "目标箱数", "出品率"}); err != nil {
		return "", err
	}
	rows := append([]CropRow(nil), report.Rows...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Crop < rows[j].Crop })
	for _, row := range rows {
		record := []string{row.Crop, strconv.Itoa(row.ReceivedKg), strconv.Itoa(row.GradeOneKg), strconv.Itoa(row.GradeTwoKg), strconv.Itoa(row.LossKg), strconv.Itoa(row.PendingKg), strconv.Itoa(row.PackedBoxes), strconv.Itoa(row.TargetBoxes), fmt.Sprintf("%.1f", row.YieldPercent)}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func ToJSON(report DailyReport) ([]byte, error) { return json.MarshalIndent(report, "", "  ") }

func SummaryLine(report DailyReport) string {
	return fmt.Sprintf("收货 %dkg，分级 %dkg，损耗 %dkg，装箱 %d 箱，停机 %d 台", report.TotalReceivedKg, report.TotalGradedKg, report.TotalLossKg, report.TotalPackedBoxes, len(report.StoppedLines))
}

func ParseCSV(input string) ([]model.InboundCommand, error) {
	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	result := make([]model.InboundCommand, 0, len(rows))
	for index, row := range rows {
		if index == 0 && len(row) > 0 && row[0] == "id" {
			continue
		}
		if len(row) < 4 {
			return nil, fmt.Errorf("row %d needs id,crop,variety,weight", index+1)
		}
		weight, parseErr := strconv.Atoi(strings.TrimSpace(row[3]))
		if parseErr != nil || weight <= 0 {
			return nil, fmt.Errorf("row %d has invalid weight", index+1)
		}
		command := model.InboundCommand{ID: strings.TrimSpace(row[0]), Crop: strings.TrimSpace(row[1]), Variety: strings.TrimSpace(row[2]), ReceivedKg: weight}
		if err := command.Validate(); err != nil {
			return nil, fmt.Errorf("row %d: %w", index+1, err)
		}
		result = append(result, command)
	}
	return result, nil
}

func WriteReport(w io.Writer, report DailyReport, format string) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		data, err := ToJSON(report)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	case "csv":
		data, err := ToCSV(report)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, data)
		return err
	case "summary", "":
		_, err := io.WriteString(w, SummaryLine(report))
		return err
	default:
		return fmt.Errorf("unknown report format %q", format)
	}
}
