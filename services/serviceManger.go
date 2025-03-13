package services

import (
	"context"
	"foo/services/registry"
	"log"
	"sync"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

type Manager struct {
	services map[string]ServiceConfig
	wg       sync.WaitGroup
	registry *registry.Registry
	ctx      context.Context
	cancel   context.CancelFunc
}
type ServiceConfig struct {
	service Service
	started bool
}

func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		services: make(map[string]ServiceConfig),
		ctx:      ctx,
		cancel:   cancel,
		registry: registry.NewRegistry(),
	}
}
func (m *Manager) GetRegistry() *registry.Registry {
	return m.registry
}

func (m *Manager) Register(service Service) {
	m.services[service.Name()] = ServiceConfig{
		service: service,
		started: false,
	}
	log.Print("Registered service: ", service.Name())
}

func (m *Manager) Start() error {
	for _, sCfg := range m.services {
		if sCfg.started {
			continue
		}
		svc := sCfg.service // Create a new variable to avoid closure issues
		m.wg.Add(1)
		go func() {

			defer m.wg.Done()
			if err := svc.Start(m.ctx); err != nil {
				log.Printf("Error starting service %s: %v", svc.Name(), err)
			}
		}()
	}
	return nil
}

func (m *Manager) StartService(service Service) error {
	m.services[service.Name()] = ServiceConfig{
		service: service,
		started: true,
	}
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := service.Start(m.ctx); err != nil {
			log.Printf("Error starting service %s: %v", service.Name(), err)
		} else {
			log.Printf("Service %s started", service.Name())
		}
	}()
	return nil
}

func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}
