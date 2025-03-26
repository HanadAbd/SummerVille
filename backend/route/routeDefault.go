package route

import (
	"encoding/json"
	"fmt"
	"foo/backend/connections"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	query := requestData.Query
	var response struct {
		Result   interface{} `json:"result"`
		Metadata struct {
			Dependencies []struct {
				ID       string      `json:"id"`
				Type     string      `json:"type"`
				Query    string      `json:"query"`
				Value    interface{} `json:"value"`
				Function string      `json:"function"`
				Depends  []string    `json:"depends_on,omitempty"`
				Order    int         `json:"order"`
			} `json:"dependencies"`
		} `json:"metadata"`
	}

	// Run the main query
	rows, err := prodConn.Conn.Query(query)
	if err != nil {
		http.Error(w, "Query execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Store the main result
	response.Result, err = connections.ScanRows(rows)
	if err != nil {
		http.Error(w, "Failed to scan results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process query for metadata
	queryLower := strings.ToLower(query)
	tableName := extractTableName(query)
	dependencyGraph := []struct {
		ID       string      `json:"id"`
		Type     string      `json:"type"`
		Query    string      `json:"query"`
		Value    interface{} `json:"value"`
		Function string      `json:"function"`
		Depends  []string    `json:"depends_on,omitempty"`
		Order    int         `json:"order"`
	}{}

	// Handle WHERE clause first (filtering always happens first)
	if strings.Contains(queryLower, "where") {
		whereInfo := extractWhereInfo(query)
		filterDep := struct {
			ID       string      `json:"id"`
			Type     string      `json:"type"`
			Query    string      `json:"query"`
			Value    interface{} `json:"value"`
			Function string      `json:"function"`
			Depends  []string    `json:"depends_on,omitempty"`
			Order    int         `json:"order"`
		}{
			ID:       "filter",
			Type:     "WHERE",
			Query:    query,
			Value:    whereInfo,
			Function: "filter_results",
			Order:    1,
		}
		dependencyGraph = append(dependencyGraph, filterDep)
	}

	// Handle COUNT (whether explicit or needed for AVG)
	needsCount := strings.Contains(queryLower, "count(") || strings.Contains(queryLower, "avg(")
	var countID string = "count"

	if needsCount {
		column := extractColumnName(query)
		countQuery := fmt.Sprintf("SELECT COUNT(%s) FROM %s", column, tableName)
		if strings.Contains(queryLower, "where") {
			// Append the where clause
			countQuery += " " + strings.Split(query, "WHERE")[1]
		}

		// Execute the count query to get the actual value
		var countValue interface{}
		countRows, err := prodConn.Conn.Query(countQuery)
		if err == nil {
			defer countRows.Close()
			countResult, _ := connections.ScanRows(countRows)
			if len(countResult.([]map[string]interface{})) > 0 {
				countValue = countResult.([]map[string]interface{})[0]
			}
		}

		countDep := struct {
			ID       string      `json:"id"`
			Type     string      `json:"type"`
			Query    string      `json:"query"`
			Value    interface{} `json:"value"`
			Function string      `json:"function"`
			Depends  []string    `json:"depends_on,omitempty"`
			Order    int         `json:"order"`
		}{
			ID:       countID,
			Type:     "COUNT",
			Query:    countQuery,
			Value:    countValue,
			Function: "increment_count",
			Order:    2,
		}

		// If there's a filter, this count depends on it
		if strings.Contains(queryLower, "where") {
			countDep.Depends = []string{"filter"}
		}

		dependencyGraph = append(dependencyGraph, countDep)
	}

	// Handle SUM (whether explicit or needed for AVG)
	needsSum := strings.Contains(queryLower, "sum(") || strings.Contains(queryLower, "avg(")
	var sumID string = "sum"

	if needsSum {
		column := extractColumnName(query)
		sumQuery := fmt.Sprintf("SELECT SUM(%s) FROM %s", column, tableName)
		if strings.Contains(queryLower, "where") {
			sumQuery += " " + strings.Split(query, "WHERE")[1]
		}

		// Execute the sum query to get the actual value
		var sumValue interface{}
		sumRows, err := prodConn.Conn.Query(sumQuery)
		if err == nil {
			defer sumRows.Close()
			sumResult, _ := connections.ScanRows(sumRows)
			if len(sumResult.([]map[string]interface{})) > 0 {
				sumValue = sumResult.([]map[string]interface{})[0]
			}
		}

		sumDep := struct {
			ID       string      `json:"id"`
			Type     string      `json:"type"`
			Query    string      `json:"query"`
			Value    interface{} `json:"value"`
			Function string      `json:"function"`
			Depends  []string    `json:"depends_on,omitempty"`
			Order    int         `json:"order"`
		}{
			ID:       sumID,
			Type:     "SUM",
			Query:    sumQuery,
			Value:    sumValue,
			Function: "add_to_sum",
			Order:    2,
		}

		// If there's a filter, this sum depends on it
		if strings.Contains(queryLower, "where") {
			sumDep.Depends = []string{"filter"}
		}

		dependencyGraph = append(dependencyGraph, sumDep)
	}

	// Handle AVG (depends on both COUNT and SUM)
	if strings.Contains(queryLower, "avg(") {
		// Generate a random 3-letter identifier for this AVG operation
		avgID := fmt.Sprintf("avg_%c%c%c",
			'a'+byte(rand.Intn(26)),
			'a'+byte(rand.Intn(26)),
			'a'+byte(rand.Intn(26)))

		column := extractColumnName(query)
		avgQuery := fmt.Sprintf("SELECT AVG(%s) FROM %s", column, tableName)
		if strings.Contains(queryLower, "where") {
			avgQuery += " " + strings.Split(query, "WHERE")[1]
		}

		// Get the actual AVG value - fix for decimal handling
		var avgValue float64
		avgRows, err := prodConn.Conn.Query(avgQuery)
		if err == nil {
			defer avgRows.Close()
			avgResult, _ := connections.ScanRows(avgRows)
			if len(avgResult.([]map[string]interface{})) > 0 {
				for _, v := range avgResult.([]map[string]interface{})[0] {
					// Try to convert to float64 based on the type
					switch val := v.(type) {
					case float64:
						avgValue = val
					case float32:
						avgValue = float64(val)
					case int:
						avgValue = float64(val)
					case int64:
						avgValue = float64(val)
					case string:
						parsedVal, parseErr := strconv.ParseFloat(val, 64)
						if parseErr == nil {
							avgValue = parsedVal
						}
					}
					break
				}
			}
		}

		avgDep := struct {
			ID       string      `json:"id"`
			Type     string      `json:"type"`
			Query    string      `json:"query"`
			Value    interface{} `json:"value"`
			Function string      `json:"function"`
			Depends  []string    `json:"depends_on,omitempty"`
			Order    int         `json:"order"`
		}{
			ID:       avgID,
			Type:     "AVG",
			Query:    avgQuery,
			Function: "divide_sum_by_count",
			Depends:  []string{sumID, countID},
			Order:    3,
		}

		// Store the float value directly without any encoding
		avgDep.Value = map[string]interface{}{
			"result": avgValue,
			"steps": []map[string]string{
				{"operation": "Get sum of all values", "uses": "SUM function"},
				{"operation": "Get count of all values", "uses": "COUNT function"},
				{"operation": "Divide sum by count", "uses": "Division operation"},
			},
		}

		dependencyGraph = append(dependencyGraph, avgDep)
	}

	// Rest of the function continues as before
	if strings.Contains(queryLower, "max(") {
		column := extractColumnName(query)
		maxQuery := fmt.Sprintf("SELECT MAX(%s) FROM %s", column, tableName)
		if strings.Contains(queryLower, "where") {
			maxQuery += " " + strings.Split(query, "WHERE")[1]
		}

		// Execute the max query
		var maxValue interface{}
		maxRows, err := prodConn.Conn.Query(maxQuery)
		if err == nil {
			defer maxRows.Close()
			maxResult, _ := connections.ScanRows(maxRows)
			if len(maxResult.([]map[string]interface{})) > 0 {
				maxValue = maxResult.([]map[string]interface{})[0]
			}
		}

		maxDep := struct {
			ID       string      `json:"id"`
			Type     string      `json:"type"`
			Query    string      `json:"query"`
			Value    interface{} `json:"value"`
			Function string      `json:"function"`
			Depends  []string    `json:"depends_on,omitempty"`
			Order    int         `json:"order"`
		}{
			ID:       "max",
			Type:     "MAX",
			Query:    maxQuery,
			Value:    maxValue,
			Function: "replace_if_greater",
			Order:    2,
		}

		if strings.Contains(queryLower, "where") {
			maxDep.Depends = []string{"filter"}
		}

		dependencyGraph = append(dependencyGraph, maxDep)
	}

	// Add a final "result" node that depends on the highest-order operations
	resultDep := struct {
		ID       string      `json:"id"`
		Type     string      `json:"type"`
		Query    string      `json:"query"`
		Value    interface{} `json:"value"`
		Function string      `json:"function"`
		Depends  []string    `json:"depends_on,omitempty"`
		Order    int         `json:"order"`
	}{
		ID:       "result",
		Type:     "RESULT",
		Query:    query,
		Value:    response.Result,
		Function: "return_result",
		Order:    10,
	}

	// Find the highest-order operations to depend on
	highestOrder := 0
	highestOrderIds := []string{}

	for _, dep := range dependencyGraph {
		if dep.Order > highestOrder {
			highestOrder = dep.Order
			highestOrderIds = []string{dep.ID}
		} else if dep.Order == highestOrder {
			highestOrderIds = append(highestOrderIds, dep.ID)
		}
	}

	resultDep.Depends = highestOrderIds
	dependencyGraph = append(dependencyGraph, resultDep)

	// Assign the dependency graph to the response
	for _, dep := range dependencyGraph {
		response.Metadata.Dependencies = append(response.Metadata.Dependencies, struct {
			ID       string      `json:"id"`
			Type     string      `json:"type"`
			Query    string      `json:"query"`
			Value    interface{} `json:"value"`
			Function string      `json:"function"`
			Depends  []string    `json:"depends_on,omitempty"`
			Order    int         `json:"order"`
		}{
			ID:       dep.ID,
			Type:     dep.Type,
			Query:    dep.Query,
			Value:    dep.Value,
			Function: dep.Function,
			Depends:  dep.Depends,
			Order:    dep.Order,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Extract WHERE clause information
func extractWhereInfo(query string) map[string]string {
	parts := strings.Split(strings.ToLower(query), "where")
	if len(parts) < 2 {
		return nil
	}

	whereClause := strings.TrimSpace(parts[1])

	// Simple parser for demonstration - would need more robust parsing in production
	field := ""
	condition := ""

	// Check for common operators
	for _, op := range []string{"=", "<>", "!=", ">=", "<=", ">", "<", "like"} {
		if idx := strings.Index(whereClause, op); idx != -1 {
			field = strings.TrimSpace(whereClause[:idx])
			condition = strings.TrimSpace(whereClause[idx:])
			break
		}
	}

	return map[string]string{
		"field":     field,
		"condition": condition,
	}
}

// Helper functions (implement these based on your needs)
func extractTableName(query string) string {
	// Implement table name extraction logic
	return ""
}

func extractColumnName(query string) string {
	// Implement column name extraction logic
	return ""
}

func HandlePostgres(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
	// postgresConn := connectors.
	// connections.WriteQuery(postgresConn, "SELECT * FROM production_data LIMIT 100", w)
}
