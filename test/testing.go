package test

import (
	"context"
	"fmt"
	"foo/services"
	"foo/services/util"
	"html/template"
	"os"

	"github.com/joho/godotenv"
)

var templates *template.Template

func IntialiseServices(servicesType ...util.ServiceType) *util.Registry {
	intaliseTemplates()
	intilaseEnv()
	config, err := util.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	manager := services.NewManager()

	registry := manager.GetRegistry()

	createService(config, registry, manager, servicesType...)

	if err := manager.Start(); err != nil {
		fmt.Printf("Error starting services: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Shutting down services...")
	manager.Stop()
	fmt.Println("Server stopped gracefully")
	return registry
}

func intaliseTemplates() {
	var err error
	templates, err = template.ParseGlob("../web/dist/templates/*.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v\n", err)
	}
}
func intilaseEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func createService(config *util.Config, registry *util.Registry, manager *services.Manager, servicesType ...util.ServiceType) {
	connectionsService := services.NewConnectionsService(config, registry)
	manager.Register(connectionsService)
	if err := connectionsService.Start(context.Background()); err != nil {
		fmt.Printf("Error starting connections service: %v\n", err)
		os.Exit(1)
	}

	for _, serviceType := range servicesType {
		switch serviceType {
		case util.BackendService:
			webService := services.NewWebService(config.ServerAddress, templates, registry)
			manager.Register(webService)
			backendService := services.NewBackendService(webService.GetMux(), registry)
			manager.Register(backendService)
		case util.WebService:
			webService := services.NewWebService(config.ServerAddress, templates, registry)
			manager.Register(webService)
		case util.ETLService:
			etlService := services.NewEtlService()
			manager.Register(etlService)
		case util.SimulateService:
			simDataService := services.NewSimulatedService(registry)
			manager.Register(simDataService)
		default:
			fmt.Println("Unknown service type")
		}
	}
}
