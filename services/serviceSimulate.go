package services

import (
	"context"
	"foo/services/util"
	"foo/simData"
	"sync"
)

type SimulatedService struct {
	dataSources map[string]*simData.DataSource
	factory     *simData.Factory
	mutex       sync.RWMutex
	wg          sync.WaitGroup
	registry    *util.Registry

	simCtx    context.Context
	simCancel context.CancelFunc
}

func NewSimulatedService(registry *util.Registry) *SimulatedService {
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

	s.simCtx, s.simCancel = context.WithCancel(context.Background())

	s.dataSources = simData.IntialiseConnections(s.registry)

	s.factory = simData.IntiliaseFactory(s.dataSources)

	s.mutex.Unlock()

	simData.SetRegistry(s.registry)

	s.registry.Register("simData.factory", s.factory)
	s.registry.Register("simData.dataSources", s.dataSources)

	go func() {
		defer s.wg.Done()

		simData.SimulateData(s.dataSources, s.factory, s.simCtx)

		<-ctx.Done()
	}()

	return nil
}

func (s *SimulatedService) Stop(ctx context.Context) error {
	s.mutex.Lock()
	if s.simCancel != nil {
		s.simCancel()
	}
	s.mutex.Unlock()

	s.wg.Wait()
	<-ctx.Done()
	return nil
}

func (s *SimulatedService) GetFactory() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.factory.GetAllNodes()
}
