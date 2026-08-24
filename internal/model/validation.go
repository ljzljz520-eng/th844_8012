package model

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidID     = errors.New("id is required")
	ErrInvalidWeight = errors.New("weight must be positive")
	ErrInvalidGrade  = errors.New("grade must be one or two")
	ErrInvalidCrop   = errors.New("crop is required")
)

func (c InboundCommand) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return ErrInvalidID
	}
	if strings.TrimSpace(c.Crop) == "" {
		return ErrInvalidCrop
	}
	if c.ReceivedKg <= 0 {
		return ErrInvalidWeight
	}
	if c.ReceivedKg > 1000000 {
		return fmt.Errorf("received weight exceeds shift limit")
	}
	return nil
}

func (c GradeCommand) Validate() error {
	if strings.TrimSpace(c.BatchID) == "" {
		return ErrInvalidID
	}
	if c.Grade != "one" && c.Grade != "two" && c.Grade != "loss" {
		return ErrInvalidGrade
	}
	if c.WeightKg <= 0 {
		return ErrInvalidWeight
	}
	if strings.TrimSpace(c.Inspector) == "" {
		return errors.New("inspector is required")
	}
	return nil
}

func (c ProgressCommand) Validate() error {
	if strings.TrimSpace(c.BatchID) == "" || strings.TrimSpace(c.LineID) == "" {
		return ErrInvalidID
	}
	if c.Boxes <= 0 {
		return errors.New("boxes must be positive")
	}
	if c.TargetBoxes < c.Boxes {
		return errors.New("target boxes must cover increment")
	}
	return nil
}

func (c StopCommand) Validate() error {
	if strings.TrimSpace(c.Actor) == "" {
		return errors.New("actor is required")
	}
	if strings.TrimSpace(c.Reason) == "" {
		return errors.New("reason is required")
	}
	return nil
}
