package services

import (
	"context"
	"foo/services/registry"
	"sync"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

type Manager struct {
	services []Service
	wg       sync.WaitGroup
	registry *registry.Registry
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:      ctx,
		cancel:   cancel,
		registry: registry.NewRegistry(),
	}
}
func (m *Manager) GetRegistry() *registry.Registry {
	return m.registry
}

func (m *Manager) Register(service Service) {
	m.services = append(m.services, service)
}

func (m *Manager) Start() error {
	for _, service := range m.services {
		svc := service // Create a new variable to avoid closure issues
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			if err := svc.Start(m.ctx); err != nil {
				// Log error or handle it as needed
			}
		}()
	}
	return nil
}

func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}
