package route

import (
	"fmt"
	"foo/backend/connections"
	"html/template"
	"net/http"
)

func WebRouting(mux *http.ServeMux, templates *template.Template) {
	mux.HandleFunc("/dashboard", HandlePage(templates, "Dashboard", "dashboard"))
	mux.HandleFunc("/etl", HandlePage(templates, "Extract Transform Load", "etlProcess"))
	mux.HandleFunc("/queryBuilder", HandlePage(templates, "Query Builder", "queryBuilder"))
	mux.HandleFunc("/api/etl", HandleNewETL(templates))
	mux.HandleFunc("/api/etl", HandleGetETL(templates))
	mux.HandleFunc("/api/data-source", HandleSetDataSources)
	mux.HandleFunc("/api/data-source", HandleGetDataSources)

}

func HandlePage(templates *template.Template, name, file string) http.HandlerFunc {
	html_file := file + ".html"
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title": name,
		}
		templates.ExecuteTemplate(w, html_file, data)
	}
}

func HandleNewETL(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title": "New ETL",
		}
		templates.ExecuteTemplate(w, "newETL.html", data)
	}
}

func HandleGetETL(templates *template.Template) http.HandlerFunc {
	query := "SELECT * FROM prod.ex"
	rows, err := connections.ProdConn.Query()
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title": "Get ETL",
		}
		templates.ExecuteTemplate(w, "getETL.html", data)
	}
}

func HandleGetDataSources(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := connections.GetDataSources()
		for _, row := range data {
			fmt.Fprintf(w, "%v\n", row)
		}

	}
}
func HandleSetDataSources(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		inserted := connections.SetDataSource(w, r)
		if inserted {
			fmt.Fprintf(w, "Data Source Inserted")
		}
	}
}
