package main

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	"foo/services"
	"foo/services/util"

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
	// doBuild()
	intaliseTemplates()
	intilaseEnv()
	config, err := util.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	manager := services.NewManager()

	util := manager.GetRegistry()

	createService(config, util, manager)

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

func createService(config *util.Config, registry *util.Registry, manager *services.Manager) {
	connectionsService := services.NewConnectionsService(config, registry)
	manager.Register(connectionsService)
	if err := connectionsService.Start(context.Background()); err != nil {
		fmt.Printf("Error starting connections service: %v\n", err)
		os.Exit(1)
	}

	webService := services.NewWebService(config.ServerAddress, templates, registry)
	manager.Register(webService)

	backendService := services.NewBackendService(webService.GetMux(), registry)
	manager.Register(backendService)

	simDataService := services.NewSimulatedService(registry)
	manager.Register(simDataService)

	etlService := services.NewEtlService()
	manager.Register(etlService)
}
