package connections

import (
	"fmt"
	"time"
)

/*
This script will manage the environment by stating how connections are defined in different enviroments
*/

type Connections interface {
	// Write operations
	AddData(table TableDefinition, data []interface{}) error
	IntialiseData(table TableDefinition) error
	PurgeAllData() error
	PurgeData(table TableDefinition) error

	// Read operations
	GetData(table TableDefinition) ([]interface{}, error)
	GetDataWithFilter(table TableDefinition, filter map[string]interface{}) ([]interface{}, error)
	CountRows(table TableDefinition) (int64, error)

	// Connection management
	RetryConnection(maxAttempts int, delay time.Duration) error
	MonitorConnection() ConnectionMetrics
	CloseConnection() error
}

type QueryResult struct {
	RowsAffected int64
	LastInsertId int64
	Duration     time.Duration
	Error        error
}

type ConnectionMetrics struct {
	OpenConnections int
	IdleConnections int
	QueryCount      int64
	LastQueryTime   time.Duration
	Status          string
	LastError       error
	LastErrorTime   time.Time
	LastInfo        string
}

type AlertConfig struct {
	MaxRetryAttempts  int
	ConnectionTimeout time.Duration
	AlertThreshold    time.Duration
	EnableEmailAlerts bool
	AlertEmail        string
}

type ColumnType string

const (
	TypeInt     ColumnType = "INT"
	TypeBigInt  ColumnType = "BIGINT"
	TypeText    ColumnType = "TEXT"
	TypeVarchar ColumnType = "VARCHAR"
	TypeDate    ColumnType = "DATE"
	TypeBoolean ColumnType = "BOOLEAN"
	TypeFloat   ColumnType = "FLOAT"
	TypeJSON    ColumnType = "JSON"
	TypeUUID    ColumnType = "UUID"
	TypeTime    ColumnType = "TIMESTAMP"
)

type ColumnDefinition struct {
	Name     string
	Type     ColumnType
	Nullable bool
	Default  interface{}
}

type TableDefinition struct {
	Name    string
	Schema  string
	Columns []ColumnDefinition
}

func (t *TableDefinition) GetTableName() string {
	return fmt.Sprintf("%s.%s", t.Schema, t.Name)
}

func (t *TableDefinition) GetColumns() []string {

	var header = make([]string, len(t.Columns))
	for index, col := range t.Columns {
		header[index] = col.Name
	}
	return header
}

type ProdConn = PostgresConn

type Connector struct {
	WorkspaceID int
	PostgresDB  map[string]*PostgresConn
	CSVfile     map[string]*CSVConn
	Kafka       map[string]*KafkaConn
}

type WorkspaceConnectors map[string]*Connector

func (w *WorkspaceConnectors) AddConnector(workspaceID int, connector *Connector) {
	(*w)[fmt.Sprintf("%d", workspaceID)] = connector
}

func (w *WorkspaceConnectors) GetConnector(workspaceID int) *Connector {
	return (*w)[fmt.Sprintf("%d", workspaceID)]
}

func (w *WorkspaceConnectors) AddData(datatype string, table TableDefinition, data []interface{}) error {
	connector := w.GetConnector(1)
	if connector == nil {
		return fmt.Errorf("connector for workspace ID 1 not found")
	}

	tableName := table.GetTableName()

	switch datatype {
	case "postgres":
		if connector.PostgresDB == nil {
			return fmt.Errorf("postgres connection for %s not found", tableName)
		}
		return connector.PostgresDB[tableName].AddData(table, data)
	case "csv":
		if connector.CSVfile[tableName] == nil {
			connector.CSVfile[tableName] = NewCSVConn(tableName)
		}
		return connector.CSVfile[tableName].AddData(table, data)
	case "kafka":
		if connector.Kafka[tableName] == nil {
			connector.Kafka[tableName] = NewKafkaConn(tableName)
		}
		return connector.Kafka[tableName].AddData(table, data)
	default:
		return fmt.Errorf("unsupported data type: %s", datatype)
	}
}

func getDefaultMetrics() ConnectionMetrics {
	return ConnectionMetrics{
		OpenConnections: 1,
		IdleConnections: 1,
		QueryCount:      0,
		LastQueryTime:   0,
		Status:          "OK",
		LastError:       nil,
		LastErrorTime:   time.Now(),
		LastInfo:        "",
	}
}
