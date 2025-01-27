package route

import (
	"encoding/json"
	"fmt"
	"foo/backend/connections"
	"net/http"
	"net/url"
)

func DataSource(w http.ResponseWriter, r *http.Request) {
	ProdConn := connections.ProdConn
	if ProdConn == nil {
		http.Error(w, "Production database not connected", http.StatusInternalServerError)
		return
	}
	connections.WriteQuery(ProdConn, "SELECT * FROM prod.data_sources", w)
}

func HandleQuery(w http.ResponseWriter, r *http.Request) {

	encodeQuery := r.URL.Query().Get("query")

	if encodeQuery == "" {
		http.Error(w, "Missing 'query' parameter", http.StatusBadRequest)
		return
	}

	decodedQuery, err := url.QueryUnescape(encodeQuery)

	if err != nil {
		http.Error(w, "Failed to decode query parameter", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("SELECT query_text FROM prod.queries WHERE query_name = '%s'", decodedQuery)
	ProdConn := connections.ProdConn

	connections.WriteQuery(ProdConn, query, w)
}
func HandleExcel(w http.ResponseWriter, r *http.Request) {
	excelConn := connections.SourcesConn.ExcelFile[0].File
	rows, err := excelConn.GetRows("Sheet1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading Excel file: %v", err), http.StatusInternalServerError)
		return
	}

	if len(rows) < 1 {
		http.Error(w, "No data found in Excel sheet", http.StatusInternalServerError)
		return
	}

	columns := rows[0]
	var results []map[string]interface{}

	for _, row := range rows[1:] {
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			if i < len(row) {
				rowMap[col] = row[i]
			} else {
				rowMap[col] = nil
			}
		}
		results = append(results, rowMap)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
func HandlePostgres(w http.ResponseWriter, r *http.Request) {
	postgresConn := connections.SourcesConn.PostgresDB[0].Conn
	connections.WriteQuery(postgresConn, "SELECT * FROM production_data LIMIT 100", w)
}

func HandleMssql(w http.ResponseWriter, r *http.Request) {
	mssqlConn := connections.SourcesConn.MssqlDB[0].Conn
	connections.WriteQuery(mssqlConn, "SELECT TOP 100 * FROM sensor_data", w)
}
