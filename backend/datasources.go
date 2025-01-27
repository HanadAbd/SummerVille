/*
Will contain the API CALLS necessary in order to connect to each of those databases:

1. for the layout of the actual structure of the site

2. Pushing all the relevant data to the data source
*/

package backend

import (
	"fmt"
	"foo/backend/connections"
	"foo/backend/etl"
	"foo/backend/route"
	"net/http"
)

func StartBackend(mux *http.ServeMux) {

	_, err := connections.InitConnector()
	if err != nil {
		fmt.Printf("Error initializing connectors: %v\n", err)
		return
	}

	restAPIRequests(mux)

	etl.Refresh()

}
func restAPIRequests(mux *http.ServeMux) {
	mux.HandleFunc("/api/data/mssql", route.HandleMssql)
	mux.HandleFunc("/api/data/postgres", route.HandlePostgres)
	mux.HandleFunc("/api/data/excel", route.HandleExcel)
	mux.HandleFunc("/api/data_sources", route.DataSource)
	mux.HandleFunc("/api/query", route.HandleQuery)
}
