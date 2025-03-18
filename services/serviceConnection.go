package services

import (
	"context"
	"encoding/json"
	"foo/backend/connections"
	"foo/services/util"
	"log"
	"sync"
)

type ConnectionsService struct {
	prodDBName          string
	prodDB              *connections.ProdConn
	workspaceConnectors connections.WorkspaceConnectors
	config              *util.Config
	mu                  sync.RWMutex
	registry            *util.Registry
}

func NewConnectionsService(c *util.Config, registry *util.Registry) *ConnectionsService {
	return &ConnectionsService{
		prodDBName:          "prodDB",
		prodDB:              nil,
		workspaceConnectors: nil,
		config:              c,
		registry:            registry,
	}
}

func (c *ConnectionsService) Name() string {
	return "ConnectionsService"
}

func (c *ConnectionsService) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	prodDB, err := connections.InitProdConnector(c.config)

	if err != nil {
		return err
	}
	c.prodDB = prodDB

	if err := c.loadWorkspaceConnections(); err != nil {
		return err
	}

	c.registry.Register("connections.prod", c.prodDB)
	c.registry.Register("connections.connectors", c.workspaceConnectors)

	return nil
}

func (c *ConnectionsService) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := connections.CloseConnector(c.config); err != nil {
		return err
	}
	return nil
}

// func (c *ConnectionsService) GetProdDB() *connections.ProdConn {
// 	return c.prodDB
// }

// func (c *ConnectionsService) GetWorkspaceConnector(workspaceID int) *connections.Connector {
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()

// 	if connector, exists := c.workspaceConnectors[workspaceID]; exists {
// 		return connector
// 	}
// 	return nil
// }

// func (c *ConnectionsService) GetAllWorkspaceConnectors() connections.WorkspaceConnectors {
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()

// 	connectorsCopy := make(connections.WorkspaceConnectors)
// 	for id, connector := range c.workspaceConnectors {
// 		connectorsCopy[id] = connector
// 	}
// 	return connectorsCopy
// }

func (c *ConnectionsService) loadWorkspaceConnections() error {

	query := "SELECT workspace_id, workspace_name FROM prod.workspaces"
	rows, err := c.prodDB.Conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}

		connector := &connections.Connector{
			WorkspaceID: id,
			PostgresDB:  make(map[string]*connections.PostgresConn),
			MssqlDB:     make(map[string]*connections.MssqlConn),
			ExcelFile:   make(map[string]*connections.ExcelConn),
			Kafka:       make(map[string]*connections.KafkaConn),
		}

		if err := c.loadConnectionsForWorkspace(id, connector); err != nil {
			return err
		}

		c.mu.Lock()
		c.workspaceConnectors[id] = connector
		c.mu.Unlock()
	}

	return nil
}

func (c *ConnectionsService) loadConnectionsForWorkspace(workspaceID int, connector *connections.Connector) error {

	query := `
			SELECT 
				ds.source_name, 
				dst.name AS source_type
			FROM 
				prod.data_sources ds
			JOIN 
				prod.data_source_types dst ON ds.data_source_type_id = dst.data_source_type_id
			WHERE 
				ds.workspace_id = $1
		`
	rows, err := c.prodDB.Conn.Query(query, workspaceID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sourceName, sourceType string

		if err := rows.Scan(&sourceName, &sourceType); err != nil {
			return err
		}

		var credentialsJSON []byte
		if c.config.Environment == "dev" {
			// Get the workspace config map
			workspaces := c.config.Connections.Workspaces

			// Look for connection details in the myCPU workspace
			if workspace, ok := workspaces["myCPU"]; ok {
				if wsMap, ok := workspace.(map[string]interface{}); ok {
					// Find the matching connection config based on sourceName
					if connConfig, exists := wsMap[sourceName]; exists {
						// Marshal the specific connection configuration
						var err error
						credentialsJSON, err = json.Marshal(connConfig)
						if err != nil {
							log.Printf("Error marshaling credentials for %s: %v", sourceName, err)
							continue
						}
					}
				}
			}
		}

		if err := connections.EstablishConnection(connector, sourceName, sourceType, credentialsJSON, c.config); err != nil {
			// Log the error but continue with other connections
			log.Printf("Error establishing connection to %s: %v", sourceName, err)
			continue
		}
	}

	return nil
}
