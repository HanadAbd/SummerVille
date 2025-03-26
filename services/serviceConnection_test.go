package services

import (
	"context"
	"fmt"
	"foo/backend/connections"
	"foo/services/util"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestConnections(t *testing.T) {
	manager := runConnectionService(t)

	reg := manager.GetRegistry()

	t.Run("TestDatabaseConnection", func(t *testing.T) {
		prodConn := testProdDB(t, reg)

		if prodConn == nil {
			t.Fatal("Failed to get valid database connection")
		}
	})
	t.Run("TestAddingData", func(t *testing.T) {
		prodConn := testProdDB(t, reg)

		table, data := createTest(t)
		err := prodConn.AddData(table, data)
		if err != nil {
			t.Errorf("Error adding data: %v", err)
		}

		rows, err := prodConn.Conn.Query(fmt.Sprintf(
			"SELECT COUNT(*) FROM %s.%s WHERE name = 'test'",
			table.Schema, table.Name))
		if err != nil {
			t.Errorf("Error verifying data: %v", err)
			return
		}
		defer rows.Close()

		var count int
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				t.Errorf("Error scanning count: %v", err)
				return
			}
		}

		if count < len(data) {
			t.Errorf("Expected at least %d rows, but found %d", len(data), count)
		} else {
			t.Logf("Successfully inserted %d rows into %s.%s", count, table.Schema, table.Name)
		}
	})

	// t.Run("TestCreatingCSV", func(t *testing.T) {
	// 	regData := "workspaceConnectors"
	// 	workspaceConnectorsObj, exists := reg.Get(regData)
	// 	if !exists {
	// 		t.Error("workspaceConnectors not found in registry")
	// 	}
	// 	if workspaceConnectorsObj == nil {
	// 		t.Errorf("%v is nil", regData)
	// 	}

	// 	connectors := workspaceConnectorsObj.(connections.WorkspaceConnectors)

	// 	if connectors.GetConnector(1) == nil {
	// 		t.Errorf("No workspace connectors found")
	// 	}

	// 	table, data := createTest(t)

	// 	err := connectors.AddData("csv", table, data)
	// 	if err != nil {
	// 		t.Errorf("Error adding data: %v", err)
	// 	}

	// })
	// t.Run("TestCreatingKafka", func(t *testing.T) {
	// 	regData := "workspaceConnectors"
	// 	workspaceConnectorsObj, exists := reg.Get(regData)
	// 	if !exists {
	// 		t.Error("workspaceConnectors not found in registry")
	// 	}
	// 	if workspaceConnectorsObj == nil {
	// 		t.Errorf("%v is nil", regData)
	// 	}

	// 	connectors := workspaceConnectorsObj.(connections.WorkspaceConnectors)

	// 	if connectors.GetConnector(1) == nil {
	// 		t.Errorf("No workspace connectors found")
	// 	}

	// 	table, data := createTest(t)

	// 	err := connectors.AddData("kafka", table, data)
	// 	if err != nil {
	// 		t.Errorf("Error adding data: %v", err)
	// 	}

	// })

	manager.Stop()

}

func intilaseEnv(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatalf("Error loading .env file %v\n", err)

	}
}

func initServiceDependencies(t *testing.T) (*util.Config, *util.Registry, *Manager) {
	intilaseEnv(t)
	config, err := util.LoadConfig()
	if err != nil {
		t.Fatalf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	manager := NewManager()
	util := manager.GetRegistry()
	return config, util, manager
}

func runConnectionService(t *testing.T) (manager *Manager) {
	config, util, manager := initServiceDependencies(t)

	connectionsService := NewConnectionsService(config, util)
	manager.Register(connectionsService)
	if err := connectionsService.Start(context.Background()); err != nil {
		t.Fatalf("Error starting connections service: %v\n", err)
		os.Exit(1)
	}
	return manager
}
func testProdDB(t *testing.T, reg *util.Registry) *connections.PostgresConn {
	dbName := "prodDB"

	prodDBObj, exists := reg.Get(dbName)
	if !exists {
		t.Fatalf("prodDB not found in registry")
	}
	if prodDBObj == nil {
		t.Fatalf("%v is nil", dbName)
	}

	prodDB := prodDBObj.(*connections.PostgresConn)

	if err := prodDB.MonitorConnection(5, 5); err != nil {
		t.Fatalf("Connection %v has failed monitoring %v", prodDB.Name, err)
	}

	prodConn := &connections.PostgresConn{
		Conn: prodDB.Conn,
	}
	return prodConn
}

func createTest(t *testing.T) (connections.TableDefinition, []interface{}) {
	header := []connections.ColumnDefinition{
		{
			Name: "id",
			Type: connections.TypeInt,
		},
		{
			Name: "name",
			Type: connections.TypeText,
		},
		{
			Name: "example_value",
			Type: connections.TypeInt,
		},
	}

	table := connections.TableDefinition{
		Name:    "test_table",
		Schema:  "test",
		Columns: header,
	}
	rows := 20
	var data []interface{}
	data = createTestRows(t, table, rows)

	switch len(data) {
	case 0:
		t.Errorf("No data created")
	case rows:
	default:
		t.Errorf("Unexpected number of rows: got %d, expected %d", len(data), rows)
	}
	return table, data
}

func createTestRows(t *testing.T, table connections.TableDefinition, rows int) []interface{} {
	var data []interface{}
	header := table.Columns
	t.Logf("header: %v", header)
	for i := 0; i < rows; i++ {
		data = append(data, map[string]interface{}{
			header[0].Name: i,
			header[1].Name: "test",
			header[2].Name: 100,
		})
	}
	return data
}
