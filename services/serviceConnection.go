package services

import (
	"context"
	"encoding/json"
	"fmt"
	"foo/backend/connections"
	"foo/services/util"
	"log"
	"os"
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

	if err := c.initConnections(); err != nil {
		log.Printf("Error initializing connections: %v", err)
	}

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

func (c *ConnectionsService) initConnections() error {
	// Read connections configuration
	data, err := os.ReadFile("connections.json")
	if err != nil {
		return fmt.Errorf("failed to read connections.json: %w", err)
	}

	var config struct {
		Workspaces map[string]map[string]map[string]interface{} `json:"workspaces"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse connections.json: %w", err)
	}

	// Process each workspace
	for workspaceName, datasources := range config.Workspaces {
		workspace := &connections.Connector{
			WorkspaceID: 1,
			PostgresDB:  make(map[string]*connections.PostgresConn),
			CSVfile:     make(map[string]*connections.CSVConn),
			Kafka:       make(map[string]*connections.KafkaConn),
		}

		// Process each data source in the workspace
		for dsName, dsConfig := range datasources {
			dsType, ok := dsConfig["type"].(string)
			if !ok {
				log.Printf("Warning: data source %s in workspace %s has no type", dsName, workspaceName)
				continue
			}

			switch dsType {
			case "postgres":
				// Initialize PostgreSQL connection
				host, _ := dsConfig["host"].(string)
				port, _ := dsConfig["port"].(float64)
				user, _ := dsConfig["user"].(string)
				password, _ := dsConfig["password"].(string)
				database, _ := dsConfig["database"].(string)
				sslmode, _ := dsConfig["sslmode"].(string)

				pgCred := &connections.PostgresCred{
					User:     user,
					Password: password,
					DBName:   database,
					Host:     host,
					Port:     fmt.Sprintf("%.0f", port),
					SSLMode:  sslmode,
				}

				db, err := connections.InitPostgresDB(pgCred, &connections.ConnectionMetrics{
					OpenConnections: 5,
					IdleConnections: 2,
					QueryCount:      0,
					LastQueryTime:   0,
					Status:          "initializing",
				})

				if err != nil {
					log.Printf("Error initializing PostgreSQL connection %s: %v", dsName, err)
					continue
				}

				workspace.PostgresDB[dsName] = &connections.PostgresConn{
					Conn: db,
					Name: dsName,
				}

			case "csv":
				// Initialize CSV connection
				filepath, _ := dsConfig["url"].(string)

				csvConn := &connections.CSVConn{
					Name:      dsName,
					FilePath:  filepath,
					Connected: false,
					Metrics: connections.ConnectionMetrics{
						Status: "initializing",
					},
				}

				if err := csvConn.InitCSV(); err != nil {
					log.Printf("Error initializing CSV connection %s: %v", dsName, err)
					continue
				}

				workspace.CSVfile[dsName] = csvConn

			case "kafka":
				// Initialize Kafka connection
				broker, _ := dsConfig["broker"].(string)
				topic, _ := dsConfig["topic"].(string)

				kafkaCredential := &connections.KafkaCredential{
					Name:   dsName,
					Broker: broker,
					Topic:  topic,
				}

				kafkaConn := &connections.KafkaConn{
					Name:       dsName,
					Connected:  false,
					Credential: kafkaCredential,
					Metrics: connections.ConnectionMetrics{
						Status: "initializing",
					},
				}

				if _, err := kafkaConn.InitKafka(kafkaCredential, connections.ConnectionMetrics{
					OpenConnections: 5,
					IdleConnections: 2,
					QueryCount:      0,
					LastQueryTime:   0,
					Status:          "initializing",
				}); err != nil {
					log.Printf("Error initializing Kafka connection %s: %v", dsName, err)
					continue
				}

				workspace.Kafka[dsName] = kafkaConn
			}
		}

		// Add the workspace to connectors
		c.workspaceConnectors.AddConnector(1, workspace)
	}

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

	for _, workspaceID := range c.workspaceConnectors {
		for name, conn := range workspaceID.PostgresDB {
			if conn == nil {
				continue
			}

			if err := conn.MonitorConnection(5, 2*time.Second); err != nil {
				log.Printf("PostgreSQL connection issue for %s: %v", name, err)
			}
		}

		for name, conn := range workspaceID.CSVfile {
			if conn == nil {
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

		for name, conn := range workspaceID.Kafka {
			if conn == nil {
				continue
			}
			if err := conn.MonitorConnection(5, 2*time.Second); err != nil {
				log.Printf("Kafka connection issue for %s: %v", name, err)
			}
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
