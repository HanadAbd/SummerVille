package services

import (
	"context"
	"foo/services/registry"
	"foo/simData"
	"sync"
)

type SimulatedService struct {
	dataSources map[string]*simData.DataSources
	factory     *simData.Factory
	mutex       sync.RWMutex
	wg          sync.WaitGroup
	registry    *registry.Registry
}

func NewSimulatedService(registry *registry.Registry) *SimulatedService {
	return &SimulatedService{
		registry: registry,
	}
}

func (s *SimulatedService) Name() string {
	return "SimulatedService"
}

func (s *SimulatedService) Start(ctx context.Context) error {
	s.wg.Add(1)
	s.mutex.Lock()
	s.dataSources = simData.IntialiseConnections()
	s.factory = simData.IntiliaseFactory()
	s.mutex.Unlock()

	s.registry.Register("simData.factory", s.factory)
	s.registry.Register("simData.dataSources", s.dataSources)
	go func() {
		defer s.wg.Done()
		simData.SimulateData(s.dataSources, s.factory)
		<-ctx.Done()
	}()

	return nil
}

func (s *SimulatedService) Stop(ctx context.Context) error {
	simData.CloseConnections(s.dataSources)
	return nil
}

func (s *SimulatedService) GetFactory() map[string]*simData.Node {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.factory.GetAllNodes()
}
