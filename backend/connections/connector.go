package connections

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"foo/services/util"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

/*
This script will get store the connection types and handle the query to interact with the connector
and return them both to be used elsewhere
*/

var reg *util.Registry

func SetRegistry(r *util.Registry) {
	reg = r
}

type Connector struct {
	WorkspaceID int
	PostgresDB  map[string]*PostgresConn
	MssqlDB     map[string]*MssqlConn
	ExcelFile   map[string]*ExcelConn
	Kafka       map[string]*KafkaConn
}

type WorkspaceConnectors map[int]*Connector

type ProdConn struct {
	Conn *sql.DB
}

type ExcelConn struct {
	File        *excelize.File
	Name        string
	Refreshtime time.Time
}

type PostgresConn struct {
	Conn        *sql.DB
	Name        string
	Refreshtime time.Time
}
type MssqlConn struct {
	Conn        *sql.DB
	Name        string
	Refreshtime time.Time
}
type KafkaConn struct {
	broker string
	topic  string
}

func handleQuery(db *sql.DB, query string) ([]map[string]interface{}, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func WriteQuery(db *sql.DB, query string, w http.ResponseWriter) {
	query = strings.TrimLeft(query, ";")

	startTime := time.Now()

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting columns: %v", err), http.StatusInternalServerError)
		return
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting column types: %v", err), http.StatusInternalServerError)
		return
	}

	var columns []map[string]string
	for i, col := range cols {
		columns = append(columns, map[string]string{
			"name": col,
			"type": colTypes[i].DatabaseTypeName(),
		})
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range values {
			columnPointers[i] = &values[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			rowMap[colName] = values[i]
		}
		results = append(results, rowMap)
	}

	executionTime := time.Since(startTime).Milliseconds()

	response := map[string]interface{}{
		"total_rows":        len(results),
		"execution_time_ms": executionTime,
		"columns":           columns,
		"results":           results,
		"query":             query,
	}

	if len(results) > 100 {
		response["results"] = results[:100]
		response["truncated"] = true
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

var (
	SourcesConn *Connector
)

func getMssqlDSN(server, database, trustedConnection string) string {
	return fmt.Sprintf("server=%s;database=%s;%s", server, database, trustedConnection)
}
func getPostgresDSN(user, password, dbname, host, port, sslmode string) string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
}

func GetProdDatabase(config *util.Config) (*ProdConn, error) {
	ProdConn := &ProdConn{
		Conn: nil,
	}
	var err error
	prodEnv := getProdCred(config)
	ProdConn.Conn, err = sql.Open("postgres", prodEnv)
	if err != nil {
		err = fmt.Errorf("ERROR: Error opening database: %w", err)
		return nil, err
	}

	ProdConn.Conn.SetMaxOpenConns(1)
	ProdConn.Conn.SetMaxIdleConns(1)
	ProdConn.Conn.SetConnMaxLifetime(0)

	err = ProdConn.Conn.Ping()
	if err != nil {

		return nil, err
	}

	return ProdConn, nil
}
func getProdCred(c *util.Config) string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		c.ProdDBUser,
		c.ProdDBPassword,
		c.ProdDBName,
		c.ProdDBHost,
		c.ProdDBPort,
		c.ProdDBSSLMode,
	)
}

func InitProdConnector(config *util.Config) (*ProdConn, error) {
	prodConn, err := GetProdDatabase(config)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := IntaliseProdDB(prodConn.Conn); err != nil {
		log.Println("Error intialising the Production Database: %q", err)
		return nil, err
	}
	return prodConn, nil
}

func IntaliseProdDB(conn *sql.DB) error {
	SourcesConn = &Connector{
		PostgresDB: make(map[string]*PostgresConn),
		MssqlDB:    make(map[string]*MssqlConn),
		ExcelFile:  make(map[string]*ExcelConn),
		Kafka:      make(map[string]*KafkaConn),
	}

	// Check if workspace table exists
	var exists bool
	err := conn.QueryRow(`
		SELECT EXISTS (
			SELECT FROM pg_tables
			WHERE schemaname = 'prod' AND tablename = 'workspaces'
		);
	`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking workspace table: %v", err)
	}

	if !exists {
		// Read and execute workspace.sql
		workspaceSQL, err := os.ReadFile("backend/initDB/workspace.sql")
		if err != nil {
			return fmt.Errorf("error reading workspace.sql: %v", err)
		}
		_, err = conn.Exec(string(workspaceSQL))
		if err != nil {
			return fmt.Errorf("error executing workspace.sql: %v", err)
		}

		// Read and execute etlPipeline.sql
		etlSQL, err := os.ReadFile("backend/initDB/etlPipeline.sql")
		if err != nil {
			return fmt.Errorf("error reading etlPipeline.sql: %v", err)
		}
		_, err = conn.Exec(string(etlSQL))
		if err != nil {
			return fmt.Errorf("error executing etlPipeline.sql: %v", err)
		}

		log.Println("Database initialized successfully")
	} else {
		log.Println("Database tables already exist, skipping initialization")
	}

	return nil
}

func EstablishConnection(connector *Connector, sourceName, sourceType string, credentialsJSON []byte, config *util.Config) error {
	var credentials map[string]string
	if err := json.Unmarshal(credentialsJSON, &credentials); err != nil {
		return err
	}

	switch sourceType {
	case "postgres":
		user := credentials["user"]
		password := credentials["password"]
		dbname := credentials["dbname"]
		host := credentials["host"]
		port := credentials["port"]
		sslmode := credentials["sslmode"]

		dsn := getPostgresDSN(user, password, dbname, host, port, sslmode)
		db, err := InitPostgresDB(dsn, 5, 2, 5*time.Minute)
		if err != nil {
			return err
		}

		connector.PostgresDB[sourceName] = &PostgresConn{
			Conn:        db,
			Name:        sourceName,
			Refreshtime: time.Now(),
		}

	case "mssql":
		server := credentials["server"]
		database := credentials["database"]
		trustedConnection := credentials["trustedConnection"]

		dsn := getMssqlDSN(server, database, trustedConnection)
		db, err := InitMssqlDB(dsn, 5, 2, 5*time.Minute)
		if err != nil {
			return err
		}

		connector.MssqlDB[sourceName] = &MssqlConn{
			Conn:        db,
			Name:        sourceName,
			Refreshtime: time.Now(),
		}

	case "excel":
		url := credentials["url"]
		file, err := InitExcel(url)
		if err != nil {
			return err
		}

		connector.ExcelFile[sourceName] = &ExcelConn{
			File:        file,
			Name:        sourceName,
			Refreshtime: time.Now(),
		}

	case "kafka":
		broker := credentials["broker"]
		topic := credentials["topic"]

		kafkaConn := InitKafka(broker, topic)
		connector.Kafka[sourceName] = kafkaConn

	default:
		return fmt.Errorf("unsupported data source type: %s", sourceType)
	}

	return nil
}

func PopulateConnections(config *util.Config) (*Connector, error) {
	// This function is kept for backward compatibility
	// For workspace-specific connections, use the ConnectionsService methods
	return nil, nil
}

func CloseConnector(config *util.Config) error {
	//TODO: Close all data based on config, if it's APP_ENV = PROD, close all connections for that, and if dev, close all connections for that
	return nil
}
