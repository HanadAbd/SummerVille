package main

import (
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	"foo/services"

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
	config := services.LoadConfig()

	manager := services.NewManager()

	registry := manager.GetRegistry()

	webService := services.NewWebService(config.ServerAddress, templates, registry)
	manager.Register(webService)

	backendService := services.NewBackendService(webService.GetMux())
	manager.Register(backendService)

	simDataService := services.NewSimulatedService(registry)
	manager.Register(simDataService)

	if err := manager.Start(); err != nil {
		fmt.Printf("Error starting services: %v\n", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down services...")
	manager.Stop()
	fmt.Println("Server stopped gracefully")

}
