package services

import (
	"context"
	"fmt"
	"foo/backend/connections"
	"foo/backend/route"
	"foo/services/util"
	"net/http"
	"sync"
)

type BackendService struct {
	mutex    sync.RWMutex
	mux      *http.ServeMux
	registry *util.Registry
}

func NewBackendService(mux *http.ServeMux, registry *util.Registry) *BackendService {
	return &BackendService{
		mux:      mux,
		registry: registry,
	}
}

func (s *BackendService) Name() string {
	return "BackendService"
}

func (s *BackendService) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	prodConnVal, ok := s.registry.Get("connections.prod")
	if !ok || prodConnVal == nil {
		return fmt.Errorf("production database connection not found in registry")
	}

	ProdConn, ok := prodConnVal.(*connections.ProdConn)
	if !ok {
		return fmt.Errorf("invalid type for production connection in registry")
	}

	connectorsVal, ok := s.registry.Get("connections.connectors")
	if !ok || connectorsVal == nil {
		return fmt.Errorf("workspace connectors not found in registry")
	}

	Connectors, ok := connectorsVal.(connections.WorkspaceConnectors)
	if !ok {
		return fmt.Errorf("invalid type for workspace connectors in registry")
	}

	// Higher-order function to create route handlers
	makeHandler := func(handlerFunc func(http.ResponseWriter, *http.Request, *connections.ProdConn, connections.WorkspaceConnectors)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			handlerFunc(w, r, ProdConn, Connectors)
		}
	}

	// Register routes using the higher-order function
	s.mux.HandleFunc("/api/data/mssql", makeHandler(route.HandleMssql))
	s.mux.HandleFunc("/api/data/postgres", makeHandler(route.HandlePostgres))
	s.mux.HandleFunc("/api/data/excel", makeHandler(route.HandleExcel))
	s.mux.HandleFunc("/api/data_sources", makeHandler(route.DataSource))
	s.mux.HandleFunc("/api/query", makeHandler(route.HandleQuery))
	s.mux.HandleFunc("/api/query/run", makeHandler(route.RunQuery))

	<-ctx.Done()
	return nil
}

func (s *BackendService) Stop(ctx context.Context) error {

	<-ctx.Done()
	return nil
}
