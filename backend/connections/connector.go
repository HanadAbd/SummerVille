package connections

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
)

/*
This script will get store the connection types and handle the query to interact with the connector
and return them both to be used elsewhere
*/

type Connector struct {
	PostgresDB []*postgresConn
	MssqlDB    []*mssqlConn
	ExcelFile  []*excelConn
	Kafka      []*kafkaConn
}

type excelConn struct {
	File        *excelize.File
	Name        string
	Refreshtime time.Time
}

type postgresConn struct {
	Conn        *sql.DB
	Name        string
	Refreshtime time.Time
}
type mssqlConn struct {
	Conn        *sql.DB
	Name        string
	Refreshtime time.Time
}
type kafkaConn struct {
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
	results, err := handleQuery(db, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func GetDataSources() [][]string {
	rows, err := ProdConn.Query("SELECT * FROM prod.sources")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var data [][]string
	for rows.Next() {
		var sourceName, sourceType string
		var credentials []byte
		var createdAt, updatedAt time.Time
		err := rows.Scan(&sourceName, &sourceType, &createdAt, &updatedAt, &credentials)
		if err != nil {
			return nil
		}
		data = append(data, []string{sourceName, sourceType, createdAt.String()})
	}

}
func SetDataSource(w http.ResponseWriter, r *http.Request) bool {
	sourceName := r.FormValue("name")
	sourceType := r.FormValue("sourceType")
	createdAt := time.Now()
	updatedAt := time.Now()
	r.ParseForm()
	if sourceType == "excel" {
		excelUrl := r.FormValue("sourceUrl")
		credentials := map[string]string{"url": excelUrl}
		credentialsJSON, err := json.Marshal(credentials)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding credentials: %v", err), http.StatusInternalServerError)
			return false
		}
		_, err = ProdConn.Query("INSERT INTO data_sources (source_name, source_type, created_at, updated_at, created_by, credentials) VALUES ($1, $2, $3, $4, $5, $6)", sourceName, sourceType, createdAt, updatedAt, "admin", credentialsJSON)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error inserting data source: %v", err), http.StatusInternalServerError)
			return false
		}
	}
	return true
}
