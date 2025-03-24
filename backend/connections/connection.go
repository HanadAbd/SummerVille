package connections

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"foo/services/util"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
This script will get store the connection types and handle the query to interact with the connector
and return them both to be used elsewhere
*/

var Reg *util.Registry

func SetRegistry(r *util.Registry) {
	Reg = r
}
func GetProdDatabase(config *util.Config) (*ProdConn, error) {
	prodCred := getProdCred(config)
	prodMetrics := &ConnectionMetrics{
		OpenConnections: 1,
		IdleConnections: 1,
		QueryCount:      0,
		LastQueryTime:   0,
	}

	conn, err := InitPostgresDB(&prodCred, prodMetrics)
	if err != nil {
		return nil, fmt.Errorf("error initializing prodDB: %v", err)
	}

	prodDb := &PostgresConn{
		Name: "prodDB",
		Conn: conn,
	}
	intalised, err := intialiseProdConn(prodDb.Conn)
	if !intalised || err != nil {
		return nil, err
	}

	if Reg != nil {
		Reg.Register("prodDB", prodDb)
	} else {
		log.Println("Warning: Registry is nil, skipping Registration of prodDB")
	}

	return prodDb, nil
}

func getProdCred(c *util.Config) PostgresCred {
	prodCred := PostgresCred{
		User:     c.ProdDBUser,
		Password: c.ProdDBPassword,
		DBName:   c.ProdDBName,
		Host:     c.ProdDBHost,
		Port:     c.ProdDBPort,
		SSLMode:  c.ProdDBSSLMode,
	}

	return prodCred
}
func intialiseProdConn(conn *sql.DB) (bool, error) {
	var exists bool
	err := conn.QueryRow(`
		SELECT EXISTS (
			SELECT FROM pg_tables
			WHERE schemaname = 'prod' AND tablename = 'workspaces'
		);
	`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking workspace table: %v", err)
	}

	if !exists {
		var workDir string
		if os.Getenv("APP_ENV") == "prod" {
			workDir = "/root/backend/initDB"
		} else {
			execPath, err := os.Executable()
			if err != nil {
				return false, fmt.Errorf("error getting executable path: %v", err)
			}
			workDir = filepath.Join(filepath.Dir(execPath), "backend//initDB")
		}

		workspaceSQL, err := os.ReadFile(filepath.Join(workDir, "workspace.sql"))
		if err != nil {
			return false, fmt.Errorf("error reading workspace.sql: %v", err)
		}
		_, err = conn.Exec(string(workspaceSQL))
		if err != nil {
			return false, fmt.Errorf("error executing workspace.sql: %v", err)
		}

		etlSQL, err := os.ReadFile(filepath.Join(workDir, "etlPipeline.sql"))
		if err != nil {
			return false, fmt.Errorf("error reading etlPipeline.sql: %v", err)
		}
		_, err = conn.Exec(string(etlSQL))
		if err != nil {
			return false, fmt.Errorf("error executing etlPipeline.sql: %v", err)
		}

		log.Println("Database initialized successfully")
		return true, nil
	}

	log.Println("Database tables already exist, skipping initialization")
	return true, nil
}

func CloseConnector(config *util.Config) error {
	//TODO: Close all data based on config, if it's APP_ENV = PROD, close all connections for that, and if dev, close all connections for that
	return nil
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
