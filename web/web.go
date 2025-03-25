package web

import (
	"encoding/json"
	"fmt"
	"foo/services/util"
	"foo/simData"
	"html/template"
	"net/http"
	"strconv"
)

func IntaliseTemplates() *template.Template {
	templates, err := template.ParseGlob("web/src/templates/*.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v\n", err)
	}
	return templates
}

func WebRouting(mux *http.ServeMux, templates *template.Template, registry *util.Registry) {

	mux.HandleFunc("/login", HandlePage(templates, "Login", "login"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mux.HandleFunc("/home", HandlePage(templates, "Home", "home"))
	mux.HandleFunc("/datasource", HandlePage(templates, "Datasource", "dataSources"))
	mux.HandleFunc("/etl", HandlePage(templates, "ETL", "etl"))

	mux.HandleFunc("/query", HandlePage(templates, "Queries", "queries"))
	mux.HandleFunc("/dashboard", HandlePage(templates, "Dashboard", "dashboard"))
	mux.HandleFunc("/test-data", HandleTestPage(templates, "Test Data", "testData", registry))
	mux.HandleFunc("/about", HandlePage(templates, "About myProject", "about"))

	mux.HandleFunc("/system", HandlePage(templates, "System", "system"))

	mux.HandleFunc("/edit/", HandleEdit(templates))

	mux.HandleFunc("/api/etl", HandleNewETL(templates))

	// http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	serveWs(hub, w, r)
	// })

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

func HandleTestPage(templates *template.Template, name, file string, reg *util.Registry) http.HandlerFunc {
	if reg == nil {
		fmt.Printf("Registry is nil")
		return func(w http.ResponseWriter, r *http.Request) {
			templates.ExecuteTemplate(w, file+".html", map[string]interface{}{
				"Title":     name,
				"AllNodes":  "{}",
				"NodeCount": "0",
			})
		}
	}

	html_file := file + ".html"

	return func(w http.ResponseWriter, r *http.Request) {
		templates = IntaliseTemplates()

		// Get factory from registry, but handle the case when it's not available
		factoryObj, ok := reg.Get("simData.factory")

		var allNodes string = "{}"
		var nodeCount string = "0"

		if ok && factoryObj != nil {
			factory := factoryObj.(*simData.Factory)
			allNodes = jsonify(factory.GetAllNodes())
			nodeCount = strconv.Itoa(len(factory.GetAllNodes()))
		} else {
			fmt.Printf("Factory not available in registry yet\n")
		}

		data := map[string]interface{}{
			"Title":     name,
			"AllNodes":  allNodes,
			"NodeCount": nodeCount,
		}
		templates.ExecuteTemplate(w, html_file, data)
	}
}
func jsonify(reports map[string]interface{}) string {
	jsonData, err := json.MarshalIndent(reports, "", "  ")

	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return ""
	}

	return string(jsonData)
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
