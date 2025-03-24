package test

import (
	"foo/backend/connections"
	"foo/services/util"
	"foo/simData"
	"os"
	"testing"
	"time"
)

var reg *util.Registry

func TestMain(m *testing.M) {
	reg = IntialiseServices()

	factory := simData.IntiliaseFactory(nil)
	reg.Register("simData.factory", factory)

	dataSources := simData.IntialiseConnections(reg)
	reg.Register("simData.dataSources", dataSources)

	code := m.Run()
	os.Exit(code)
}

func TestFactoryWasCreated(t *testing.T) {
	t.Log("Testing factory creation")
	factory := getFactory(t)

	if count := factory.GetCount(); count < 0 {
		t.Errorf("No Nodes intilasied within factory: %d", count)
	}
}

func getFactory(t *testing.T) *simData.Factory {
	factoryObj, ok := reg.Get("simData.factory")
	var factory *simData.Factory
	if !ok && factoryObj == nil {
		t.Error("Factory not found")
	}
	factory = factoryObj.(*simData.Factory)
	t.Log("Factory found")
	return factory
}

func TestCreatingNode(t *testing.T) {
	factory := getFactory(t)
	queueSize := 100
	factory.AddNode("test_start", simData.Start{Name: "test_start"}, nil, 0, queueSize)
	factory.AddNode("test_complete", simData.Complete{Name: "test_complete"}, nil, 0, queueSize)
	factory.AddNode("test_cutting1", simData.CuttingMachine{Name: "test_cutting1"}, nil, 2*time.Second, queueSize)
	factory.AddNode("test_sensor1", simData.Sensor{Name: "test_sensor1"}, nil, 3*time.Second, queueSize)

	factory.AddEdges("test_start", "test_cutting1")
	factory.AddEdges("test_cutting1", "test_sensor1")
	factory.AddEdges("test_sensor1", "test_complete")

	if count := factory.GetCount(); count < 4 {
		t.Errorf("Nodes not added correctly, expected at least 4 but got: %d", count)
	}

	startNode := factory.GetNode("test_start")
	if startNode == nil {
		t.Fatal("Start node not found")
	}

	if len(startNode.NextNodes) != 1 {
		t.Errorf("Start node should have 1 next node, but has %d", len(startNode.NextNodes))
	}

	t.Log("Nodes added correctly")
}

func TestDataSourcesInitialization(t *testing.T) {
	connectors := GetConnector(t)

	table, data := createTest(t)

	err := connectors.AddData("csv", table, data)
	if err != nil {
		t.Errorf("Error adding data: %v", err)
	}
}

func GetConnector(t *testing.T) connections.WorkspaceConnectors {
	regData := "workspaceConnectors"
	workspaceConnectorsObj, exists := reg.Get(regData)
	if !exists {
		t.Error("workspaceConnectors not found in registry")
	}
	if workspaceConnectorsObj == nil {
		t.Errorf("%v is nil", regData)
	}

	connectors := workspaceConnectorsObj.(connections.WorkspaceConnectors)

	return connectors
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
