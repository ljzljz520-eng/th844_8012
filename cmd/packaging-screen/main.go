package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"agri-packaging/internal/audit"
	"agri-packaging/internal/catalog"
	"agri-packaging/internal/dashboard"
	"agri-packaging/internal/grading"
	"agri-packaging/internal/httpapi"
	"agri-packaging/internal/line"
	"agri-packaging/internal/packing"
	"agri-packaging/internal/planning"
	"agri-packaging/internal/report"
	"agri-packaging/internal/shift"
	"agri-packaging/internal/storage"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "./data/packaging.db", "embedded database path")
	flag.Parse()
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatal(err)
	}
	store, err := storage.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	catalogService := catalog.New(store, nil)
	lineService := line.New(store, nil)
	gradingService := grading.New(store, catalogService, nil)
	packingService := packing.New(store, lineService, nil)
	dashboardService := dashboard.New(store, nil)
	auditService := audit.New(store, nil)
	shiftService := shift.New(store, nil)
	reportService := report.New(store, nil)
	planningService := planning.New(store)
	_ = planningService.SetCapacity(planning.LineCapacity{LineID: "line-a", BoxesPerHour: 100, Priority: 1})
	services := httpapi.Services{Catalog: catalogService, Line: lineService, Grading: gradingService, Packing: packingService, Dashboard: dashboardService, Audit: auditService, Shift: shiftService, Report: reportService, Planning: planningService}
	server := httpapi.New(services, log.Default())
	log.Printf("agri packaging screen listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
