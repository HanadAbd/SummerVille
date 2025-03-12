package services

import (
	"context"
	"foo/services/registry"
	"foo/web"

	"html/template"
	"net/http"
	"time"
)

type WebService struct {
	mux       *http.ServeMux
	templates *template.Template
	server    *http.Server
	addr      string
	registry  *registry.Registry
}

func NewWebService(addr string, templates *template.Template, registry *registry.Registry) *WebService {
	mux := http.NewServeMux()
	return &WebService{
		mux:       mux,
		templates: templates,
		addr:      addr,
		registry:  registry,
	}
}

func (s *WebService) Name() string {
	return "WebService"
}

func (s *WebService) Start(ctx context.Context) error {
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/dist"))))
	web.WebRouting(s.mux, s.templates, s.registry)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
}

func (s *WebService) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *WebService) GetMux() *http.ServeMux {
	return s.mux
}
