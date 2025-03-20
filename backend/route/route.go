package route

import (
	"encoding/json"
	"fmt"
	"foo/backend/connections"
	"net/http"
	"net/url"
)

/*
This script will handle the routes for the API
*/

func DataSource(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
	ProdConn := prodConn
	if ProdConn == nil {
		http.Error(w, "Production database not connected", http.StatusInternalServerError)
		return
	}
	connections.WriteQuery(ProdConn.Conn, "SELECT * FROM prod.data_sources", w)
}

func HandleQuery(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {

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

	connections.WriteQuery(prodConn.Conn, query, w)
}

func RunQuery(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	if requestData.Query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	ProdConn := prodConn.Conn
	connections.WriteQuery(ProdConn, requestData.Query, w)
}

// func HandleExcel(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
// 	excelConn := connectors[0].ExcelFile["test"].File

// 	if excelConn == nil {
// 		http.Error(w, "Excel file not connected", http.StatusInternalServerError)
// 		return
// 	}
// 	rows, err := excelConn.GetRows("Sheet1")
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error reading Excel file: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	if len(rows) < 1 {
// 		http.Error(w, "No data found in Excel sheet", http.StatusInternalServerError)
// 		return
// 	}

// 	columns := rows[0]
// 	var results []map[string]interface{}

// 	for _, row := range rows[1:] {
// 		rowMap := make(map[string]interface{})
// 		for i, col := range columns {
// 			if i < len(row) {
// 				rowMap[col] = row[i]
// 			} else {
// 				rowMap[col] = nil
// 			}
// 		}
// 		results = append(results, rowMap)
// 	}

//		w.Header().Set("Content-Type", "application/json")
//		if err := json.NewEncoder(w).Encode(results); err != nil {
//			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
//		}
//	}
func HandlePostgres(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
	// postgresConn := connectors.
	// connections.WriteQuery(postgresConn, "SELECT * FROM production_data LIMIT 100", w)
}

// func HandleMssql(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
// 	mssqlConn := connectors[0].MssqlDB["test"].Conn
// 	connections.WriteQuery(mssqlConn, "SELECT TOP 100 * FROM sensor_data", w)
// }
