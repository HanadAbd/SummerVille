package route

import (
	"encoding/json"
	"foo/services/util"
	"foo/simData"
	"net/http"
)

/*
This script will handle the routes for the API
*/

var Reg *util.Registry

func SetRegistry(registry *util.Registry) {
	Reg = registry
}

func writeJSONResponse(w http.ResponseWriter, status int, message string, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	response := map[string]interface{}{
		"message": message,
		"status":  status,
		"payload": payload,
	}

	json.NewEncoder(w).Encode(response)
}

func writeJSONErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	response := map[string]interface{}{
		"message": message,
		"status":  status,
	}

	json.NewEncoder(w).Encode(response)
}

func getFactory() *simData.Factory {
	factoryObj, ok := Reg.Get("simData.factory")
	var factory *simData.Factory
	if !ok && factoryObj == nil {
		return nil
	}
	factory = factoryObj.(*simData.Factory)
	return factory
}
