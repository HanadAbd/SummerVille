package route

import (
	"fmt"
	"html/template"
	"net/http"
)

func IntaliseTemplates() *template.Template {
	templates, err := template.ParseGlob("web/src/templates/*.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v\n", err)
	}
	return templates
}

func WebRouting(mux *http.ServeMux, templates *template.Template) {

	mux.HandleFunc("/login", HandlePage(templates, "Login", "login"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// <a href="/home">Home</a>
	// <a href="/datasource">Data Sources</a>
	// <a href="/etl">ETL</a>
	// <a href="/query">Queries</a>
	// <a href="/dashboard">Dashboard Example</a>
	// <a href="/about">About</a>

	mux.HandleFunc("/home", HandlePage(templates, "Home", "home"))
	mux.HandleFunc("/datasource", HandlePage(templates, "Datasource", "datasources"))
	mux.HandleFunc("/etl", HandlePage(templates, "ETL", "etl"))

	mux.HandleFunc("/query", HandlePage(templates, "Queries", "queries"))
	mux.HandleFunc("/dashboard", HandlePage(templates, "Dashboard", "dashboard"))
	mux.HandleFunc("/test-data", HandlePage(templates, "Test Data", "testData"))
	mux.HandleFunc("/about", HandlePage(templates, "About myProject", "about"))

	mux.HandleFunc("/edit/", HandleEdit(templates))

	mux.HandleFunc("/api/etl", HandleNewETL(templates))

}
func HandleEdit(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates = IntaliseTemplates()
		path := r.URL.Path[len("/edit/"):]
		data := map[string]interface{}{
			"Title": "Edit " + path,
		}
		templates.ExecuteTemplate(w, "edit.html", data)
	}
}

func HandlePage(templates *template.Template, name, file string) http.HandlerFunc {
	html_file := file + ".html"
	return func(w http.ResponseWriter, r *http.Request) {
		templates = IntaliseTemplates()
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

// func HandleGetETL(templates *template.Template) http.HandlerFunc {
// 	query := "SELECT * FROM prod.ex"
// 	rows, err := connections.ProdConn.Query()
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		data := map[string]interface{}{
// 			"Title": "Get ETL",
// 		}
// 		templates.ExecuteTemplate(w, "getETL.html", data)
// 	}
// }

// func HandleGetDataSources(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "GET" {
// 		data := connections.GetDataSources()
// 		for _, row := range data {
// 			fmt.Fprintf(w, "%v\n", row)
// 		}

// 	}
// }
// func HandleSetDataSources(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "POST" {
// 		r.ParseForm()
// 		inserted := connections.SetDataSource(w, r)
// 		if inserted {
// 			fmt.Fprintf(w, "Data Source Inserted")
// 		}
// 	}
// }
