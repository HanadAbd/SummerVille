package web

import (
	"fmt"
	"foo/services/registry"
	"foo/simData"
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

func WebRouting(mux *http.ServeMux, templates *template.Template, registry interface{}) {

	mux.HandleFunc("/login", HandlePage(templates, "Login", "login"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mux.HandleFunc("/home", HandlePage(templates, "Home", "home"))
	mux.HandleFunc("/datasource", HandlePage(templates, "Datasource", "datasources"))
	mux.HandleFunc("/etl", HandlePage(templates, "ETL", "etl"))

	mux.HandleFunc("/query", HandlePage(templates, "Queries", "queries"))
	mux.HandleFunc("/dashboard", HandlePage(templates, "Dashboard", "dashboard"))
	mux.HandleFunc("/test-data", HandleTestPage(templates, "Test Data", "testData", registry))
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

func HandleTestPage(templates *template.Template, name, file string, register interface{}) http.HandlerFunc {
	html_file := file + ".html"

	reg := register.(*registry.Registry)

	factoryObj, ok := reg.Get("simData.factory")
	if !ok {
		fmt.Printf("Error getting factory object from registry")
		return nil
	}
	factory := factoryObj.(*simData.Factory)

	return func(w http.ResponseWriter, r *http.Request) {
		templates = IntaliseTemplates()
		data := map[string]interface{}{
			"Title":     name,
			"AllNodes":  factory.GetAllNodes(),
			"NodeCount": factory.GetCount(),
		}
		templates.ExecuteTemplate(w, html_file, data)
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
