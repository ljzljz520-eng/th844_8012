package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"agri-packaging/internal/audit"
	"agri-packaging/internal/catalog"
	"agri-packaging/internal/dashboard"
	"agri-packaging/internal/grading"
	"agri-packaging/internal/line"
	"agri-packaging/internal/model"
	"agri-packaging/internal/packing"
	"agri-packaging/internal/planning"
	"agri-packaging/internal/report"
	"agri-packaging/internal/shift"
)

type Services struct {
	Catalog   *catalog.Service
	Line      *line.Service
	Grading   *grading.Service
	Packing   *packing.Service
	Dashboard *dashboard.Service
	Audit     *audit.Service
	Shift     *shift.Service
	Report    *report.Service
	Planning  *planning.Service
}

type Server struct {
	services Services
	logger   *log.Logger
}

func New(services Services, logger *log.Logger) *Server {
	return &Server{services: services, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/batches", s.handleBatches)
	mux.HandleFunc("/api/grades", s.handleGrades)
	mux.HandleFunc("/api/progress", s.handleProgress)
	mux.HandleFunc("/api/lines", s.handleLines)
	mux.HandleFunc("/api/dashboard", s.handleDashboard)
	mux.HandleFunc("/api/audits", s.handleAudits)
	mux.HandleFunc("/api/shifts", s.handleShifts)
	mux.HandleFunc("/api/report", s.handleReport)
	mux.HandleFunc("/api/plan", s.handlePlan)
	return withJSON(mux)
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

func decodeJSON(r *http.Request, target any) error { return json.NewDecoder(r.Body).Decode(target) }

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "agri-packaging"})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexPage))
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	if s.logger != nil {
		s.logger.Printf("request failed: %v", err)
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}

func routeMethod(r *http.Request, allowed ...string) bool {
	for _, method := range allowed {
		if r.Method == method {
			return true
		}
	}
	return false
}

func readActor(r *http.Request) string {
	actor := r.Header.Get("X-Operator")
	if actor == "" {
		actor = "screen"
	}
	return actor
}

var _ model.ProgressCommand
