package httpapi

import (
	"net/http"
	"strconv"

	"agri-packaging/internal/model"
	"agri-packaging/internal/planning"
)

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	if !routeMethod(r, http.MethodGet) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.services.Report == nil {
		s.fail(w, model.ErrInvalidID)
		return
	}
	result, err := s.services.Report.Build(r.URL.Query().Get("shift"))
	if err != nil {
		s.fail(w, err)
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		writeJSON(w, http.StatusOK, result)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if err := s.services.Report.Write(w, result, format); err != nil {
		s.fail(w, err)
		return
	}
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	if !routeMethod(r, http.MethodPost) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.services.Planning == nil {
		s.fail(w, model.ErrInvalidID)
		return
	}
	var request struct {
		Crop           string `json:"crop"`
		PendingKg      int    `json:"pending_kg"`
		BoxWeightKg    int    `json:"box_weight_kg"`
		HoursAvailable int    `json:"hours_available"`
	}
	if err := decodeJSON(r, &request); err != nil {
		s.fail(w, err)
		return
	}
	plan, err := s.services.Planning.Build(planning.PlanRequest{Crop: request.Crop, PendingKg: request.PendingKg, BoxWeightKg: request.BoxWeightKg, HoursAvailable: request.HoursAvailable})
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (s *Server) handleBatches(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		items, err := s.services.Catalog.List()
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var command model.InboundCommand
	if err := decodeJSON(r, &command); err != nil {
		s.fail(w, err)
		return
	}
	item, err := s.services.Catalog.Receive(command)
	if err != nil {
		s.fail(w, err)
		return
	}
	_, _ = s.services.Audit.Record("receive", readActor(r), "ProduceBatch", item.ID, item.Crop)
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleGrades(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		items, err := s.services.Grading.List()
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var command model.GradeCommand
	if err := decodeJSON(r, &command); err != nil {
		s.fail(w, err)
		return
	}
	item, err := s.services.Grading.Record(command)
	if err != nil {
		s.fail(w, err)
		return
	}
	_, _ = s.services.Audit.Record("grade", readActor(r), "GradeRecord", item.ID, item.Grade)
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		items, err := s.services.Packing.List()
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var command model.ProgressCommand
	if err := decodeJSON(r, &command); err != nil {
		s.fail(w, err)
		return
	}
	item, err := s.services.Packing.Submit(command)
	if err != nil {
		s.fail(w, err)
		return
	}
	_, _ = s.services.Audit.Record("pack", readActor(r), "PackingProgress", item.ID, strconv.Itoa(command.Boxes))
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleLines(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		items, err := s.services.Line.List()
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var command struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &command); err != nil {
		s.fail(w, err)
		return
	}
	if command.Action == "register" {
		item, err := s.services.Line.Register(command.ID, command.Name)
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
		return
	}
	if command.Action == "stop" {
		item, err := s.services.Line.Stop(command.ID, model.StopCommand{Reason: command.Reason, Actor: readActor(r)})
		if err != nil {
			s.fail(w, err)
			return
		}
		_, _ = s.services.Audit.Record("stop", readActor(r), "DeviceLine", item.ID, item.Reason)
		writeJSON(w, http.StatusOK, item)
		return
	}
	if command.Action == "resume" {
		item, err := s.services.Line.Resume(command.ID, readActor(r))
		if err != nil {
			s.fail(w, err)
			return
		}
		_, _ = s.services.Audit.Record("resume", readActor(r), "DeviceLine", item.ID, "")
		writeJSON(w, http.StatusOK, item)
		return
	}
	s.fail(w, model.ErrInvalidID)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !routeMethod(r, http.MethodGet) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	snapshot, err := s.services.Dashboard.Snapshot(r.URL.Query().Get("crop"))
	if err != nil {
		s.fail(w, err)
		return
	}
	audits, err := s.services.Dashboard.RecentAudits(12)
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"snapshot": snapshot, "audits": audits})
}

func (s *Server) handleAudits(w http.ResponseWriter, r *http.Request) {
	if !routeMethod(r, http.MethodGet) {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	items, err := s.services.Audit.List()
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleShifts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var command struct {
		ID     string `json:"id"`
		Action string `json:"action"`
		Leader string `json:"leader"`
		Notes  string `json:"notes"`
	}
	if err := decodeJSON(r, &command); err != nil {
		s.fail(w, err)
		return
	}
	if command.Action == "start" {
		item, err := s.services.Shift.Start(command.ID, command.Leader)
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
		return
	}
	if command.Action == "end" {
		item, err := s.services.Shift.End(command.ID, command.Notes)
		if err != nil {
			s.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	s.fail(w, model.ErrInvalidID)
}
