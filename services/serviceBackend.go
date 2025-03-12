package services

import (
	"context"
	"foo/backend/connections"
	"foo/backend/route"
	"net/http"
)

type BackendService struct {
	mux *http.ServeMux
}

func NewBackendService(mux *http.ServeMux) *BackendService {
	return &BackendService{
		mux: mux,
	}
}

func (s *BackendService) Name() string {
	return "BackendService"
}

func (s *BackendService) Start(ctx context.Context) error {
	_, err := connections.InitConnector()
	if err != nil {
		return err
	}

	s.mux.HandleFunc("/api/data/mssql", route.HandleMssql)
	s.mux.HandleFunc("/api/data/postgres", route.HandlePostgres)
	s.mux.HandleFunc("/api/data/excel", route.HandleExcel)
	s.mux.HandleFunc("/api/data_sources", route.DataSource)
	s.mux.HandleFunc("/api/query", route.HandleQuery)
	s.mux.HandleFunc("/api/query/run", route.RunQuery)

	<-ctx.Done()
	return nil
}

func (s *BackendService) Stop(ctx context.Context) error {
	if err := connections.CloseConnector(); err != nil {
		return err
	}
	return nil
}
