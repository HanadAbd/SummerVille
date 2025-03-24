package route

import (
	"encoding/json"
	"fmt"
	"foo/backend/connections"
	"net/http"
	"net/url"
)

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

func HandlePostgres(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
	// postgresConn := connectors.
	// connections.WriteQuery(postgresConn, "SELECT * FROM production_data LIMIT 100", w)
}
