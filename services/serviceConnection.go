package services

import (
	"context"
	"foo/backend/connections"
)

type ConnectionsService struct {
	prodDBName string
	prodDB     *connections.Connector
}

func NewConnectionsService() *ConnectionsService {
	return &ConnectionsService{}
}

func (c *ConnectionsService) Name() string {
	return "ConnectionsService"
}

func (c *ConnectionsService) Start(ctx context.Context) error {

	connections.InitConnector()
	return nil
}

func (c *ConnectionsService) Stop(ctx context.Context) error {
	return nil
}
