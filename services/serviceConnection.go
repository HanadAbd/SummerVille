package services

import (
	"context"
	"foo/backend/connections"
	"foo/services/util"
	"log"
	"sync"
	"time"
)

type ConnectionsService struct {
	prodDBName          string
	prodDB              *connections.PostgresConn
	workspaceConnectors connections.WorkspaceConnectors
	config              *util.Config
	mu                  sync.RWMutex
	registry            *util.Registry

	monitorCtx    context.Context
	monitorCancel context.CancelFunc
	monitorWg     sync.WaitGroup
}

func NewConnectionsService(c *util.Config, registry *util.Registry) *ConnectionsService {
	return &ConnectionsService{
		prodDBName:          "prodDB",
		prodDB:              nil,
		workspaceConnectors: make(map[string]*connections.Connector),
		config:              c,
		registry:            registry,
	}
}

func (c *ConnectionsService) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error

	c.monitorCtx, c.monitorCancel = context.WithCancel(context.Background())
	connections.SetRegistry(c.registry)

	c.prodDB, err = connections.GetProdDatabase(c.config)
	if err != nil {
		log.Fatalf("Error intialising prodDB: %q\n", err)
	}

	if len(c.workspaceConnectors) == 0 {
		c.workspaceConnectors.AddConnector(1, &connections.Connector{
			WorkspaceID: 1,
			PostgresDB:  make(map[string]*connections.PostgresConn),
			CSVfile:     make(map[string]*connections.CSVConn),
			Kafka:       make(map[string]*connections.KafkaConn),
		})
	}

	c.registry.Register("workspaceConnectors", c.workspaceConnectors)

	c.startConnectionMonitor()

	return nil
}

func (c *ConnectionsService) startConnectionMonitor() {
	c.monitorWg.Add(1)
	go func() {
		defer c.monitorWg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		log.Println("Connection monitor started")

		for {
			select {
			case <-ticker.C:
				c.monitorConnections()
			case <-c.monitorCtx.Done():
				log.Println("Connection monitor stopping")
				return
			}
		}
	}()
}
func (c *ConnectionsService) monitorConnections() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// First check if the prodDB is healthy
	if c.prodDB != nil && c.prodDB.Conn != nil {
		if err := c.prodDB.MonitorConnection(5, 2*time.Second); err != nil {
			log.Printf("Production database connection issue: %v", err)
		}
	}

	for _, workspace := range c.workspaceConnectors {
		if workspace == nil {
			log.Printf("Warning: Found nil workspace connector")
			continue
		}

		// Monitor PostgreSQL connections
		for name, conn := range workspace.PostgresDB {
			if conn == nil || conn.Conn == nil {
				log.Printf("Warning: PostgreSQL connection %s is nil", name)
				continue
			}

			if err := conn.MonitorConnection(5, 2*time.Second); err != nil {
				log.Printf("PostgreSQL connection issue for %s: %v", name, err)
			}
		}

		// Monitor CSV connections
		for name, conn := range workspace.CSVfile {
			if conn == nil {
				log.Printf("Warning: CSV connection %s is nil", name)
				continue
			}

			metrics := conn.MonitorConnection()
			if metrics.Status != "connected" {
				log.Printf("CSV connection issue for %s: %v", name, metrics.LastError)
				if err := conn.RetryConnection(5, 2*time.Second); err != nil {
					log.Printf("Failed to reconnect CSV %s: %v", name, err)
				}
			}
		}

		// Monitor Kafka connections
		for name, conn := range workspace.Kafka {
			if conn == nil {
				log.Printf("Warning: Kafka connection %s is nil", name)
				continue
			}

			if err := conn.MonitorConnection(5, 2*time.Second); err != nil {
				log.Printf("Kafka connection issue for %s: %v", name, err)
			}
		}
	}
}

// Helper functions to monitor each type of connection
func monitorPostgresConnections(connections map[string]*connections.PostgresConn) {
	if connections == nil {
		return
	}

	for name, conn := range connections {
		if conn == nil || conn.Conn == nil {
			log.Printf("Warning: PostgreSQL connection %s is nil", name)
			continue
		}

		if err := conn.MonitorConnection(5, 2*time.Second); err != nil {
			log.Printf("PostgreSQL connection issue for %s: %v", name, err)
		}
	}
}

func monitorCSVConnections(connections map[string]*connections.CSVConn) {
	if connections == nil {
		return
	}

	for name, conn := range connections {
		if conn == nil {
			log.Printf("Warning: CSV connection %s is nil", name)
			continue
		}

		metrics := conn.MonitorConnection()
		if metrics.Status != "connected" {
			log.Printf("CSV connection issue for %s: %v", name, metrics.LastError)
			if err := conn.RetryConnection(5, 2*time.Second); err != nil {
				log.Printf("Failed to reconnect CSV %s: %v", name, err)
			}
		}
	}
}

func monitorKafkaConnections(connections map[string]*connections.KafkaConn) {
	if connections == nil {
		return
	}

	for name, conn := range connections {
		if conn == nil {
			log.Printf("Warning: Kafka connection %s is nil", name)
			continue
		}

		if err := conn.MonitorConnection(5, 2*time.Second); err != nil {
			log.Printf("Kafka connection issue for %s: %v", name, err)
		}
	}
}

func (c *ConnectionsService) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.monitorCancel != nil {
		c.monitorCancel()
		c.monitorWg.Wait()
	}

	for _, workspace := range c.workspaceConnectors {

		closeConnections(workspace.PostgresDB, "PostgreSQL")
		closeConnections(workspace.CSVfile, "CSV")
		closeConnections(workspace.Kafka, "Kafka")
	}

	if c.prodDB != nil && c.prodDB.Conn != nil {
		if err := c.prodDB.CloseConnection(); err != nil {
			log.Printf("Error closing production database: %v", err)
		}
	}

	return nil
}

func closeConnections(connections interface{}, connType string) {
	if m, ok := connections.(map[string]interface{ CloseConnection() error }); ok {
		for name, conn := range m {
			if conn != nil {
				if err := conn.CloseConnection(); err != nil {
					log.Printf("Error closing %s connection %s: %v", connType, name, err)
				}
			}
		}
	}
}

func (c *ConnectionsService) Name() string {
	return "ConnectionsService"
}
