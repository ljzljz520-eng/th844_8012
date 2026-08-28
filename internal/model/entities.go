package model

import "time"

type ProduceBatch struct {
	ID         string    `json:"id"`
	Crop       string    `json:"crop"`
	Variety    string    `json:"variety"`
	ReceivedKg int       `json:"received_kg"`
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status"`
}

type GradeRecord struct {
	ID        string    `json:"id"`
	BatchID   string    `json:"batch_id"`
	Crop      string    `json:"crop"`
	Grade     string    `json:"grade"`
	WeightKg  int       `json:"weight_kg"`
	Inspector string    `json:"inspector"`
	CreatedAt time.Time `json:"created_at"`
}

type PackingProgress struct {
	ID          string    `json:"id"`
	BatchID     string    `json:"batch_id"`
	LineID      string    `json:"line_id"`
	Boxes       int       `json:"boxes"`
	TargetBoxes int       `json:"target_boxes"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeviceLine struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Stopped   bool      `json:"stopped"`
	Reason    string    `json:"reason"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuditEvent struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type DashboardSnapshot struct {
	Crop         string    `json:"crop"`
	PendingKg    int       `json:"pending_kg"`
	GradeOneKg   int       `json:"grade_one_kg"`
	GradeTwoKg   int       `json:"grade_two_kg"`
	LossKg       int       `json:"loss_kg"`
	PackedBoxes  int       `json:"packed_boxes"`
	TargetBoxes  int       `json:"target_boxes"`
	StoppedLines int       `json:"stopped_lines"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OperatorSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ShiftSummary struct {
	ShiftID   string     `json:"shift_id"`
	Leader    string     `json:"leader"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Notes     string     `json:"notes"`
}

type ProgressCommand struct {
	BatchID     string `json:"batch_id"`
	LineID      string `json:"line_id"`
	Boxes       int    `json:"boxes"`
	TargetBoxes int    `json:"target_boxes"`
}

type GradeCommand struct {
	BatchID   string `json:"batch_id"`
	Grade     string `json:"grade"`
	WeightKg  int    `json:"weight_kg"`
	Inspector string `json:"inspector"`
}

type StopCommand struct {
	Reason string `json:"reason"`
	Actor  string `json:"actor"`
}

type InboundCommand struct {
	ID         string `json:"id"`
	Crop       string `json:"crop"`
	Variety    string `json:"variety"`
	ReceivedKg int    `json:"received_kg"`
}
