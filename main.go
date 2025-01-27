package main

import (
	"fmt"
	"html/template"
	"net/http"

	"foo/web/route"

	"github.com/joho/godotenv"
)

var templates *template.Template

func intaliseTemplates() {
	var err error
	templates, err = template.ParseGlob("web/dist/templates/*.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v\n", err)
	}
}
func intilaseEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {
	doBuild()
	intaliseTemplates()

	intilaseEnv()
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/dist"))))

	route.WebRouting(mux, templates)

	// backend.StartBackend(mux)

	fmt.Println("Server is running on: http://localhost:8080/dashboard")

	err := http.ListenAndServe("localhost:8080", mux)

	if err != nil {
		fmt.Printf("Server Crashed: %v\n", err)
	}
}
