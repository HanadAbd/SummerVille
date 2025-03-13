package services

import "context"

type ConnectionsService struct {
}

func NewConnectionsService() *ConnectionsService {
	return &ConnectionsService{}
}

func (c *ConnectionsService) Name() string {
	return "ConnectionsService"
}

func (c *ConnectionsService) Start(ctx context.Context) error {
	return nil
}

func (c *ConnectionsService) Stop(ctx context.Context) error {
	return nil
}
