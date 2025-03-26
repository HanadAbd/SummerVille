package test

// import (
// 	"foo/backend/connections"
// 	"foo/services/util"
// 	"foo/simData"
// 	"os"
// 	"testing"
// 	"time"
// )

// var reg *util.Registry

// func TestMain(m *testing.M) {
// 	reg = IntialiseServices()

// 	factory := simData.IntiliaseFactory(nil)
// 	reg.Register("simData.factory", factory)

// 	dataSources := simData.IntialiseConnections(reg)
// 	reg.Register("simData.dataSources", dataSources)

// 	code := m.Run()
// 	os.Exit(code)
// }

// func TestFactoryWasCreated(t *testing.T) {
// 	t.Log("Testing factory creation")
// 	factory := getFactory(t)

// 	if count := factory.GetCount(); count < 0 {
// 		t.Errorf("No Nodes intilasied within factory: %d", count)
// 	}
// }

// func getFactory(t *testing.T) *simData.Factory {
// 	factoryObj, ok := reg.Get("simData.factory")
// 	var factory *simData.Factory
// 	if !ok && factoryObj == nil {
// 		t.Error("Factory not found")
// 	}
// 	factory = factoryObj.(*simData.Factory)
// 	t.Log("Factory found")
// 	return factory
// }

// func TestCreatingNode(t *testing.T) {
// 	factory := getFactory(t)
// 	queueSize := 100
// 	factory.AddNode("start", &simData.Start{Name: "start"}, nil, 0, queueSize)
// 	factory.AddNode("complete", &simData.Complete{Name: "complete"}, nil, 0, queueSize)

// 	factory.AddNode("cutting1", &simData.CuttingMachineNode{
// 		Node:                simData.Node{ID: "cutting1", NodeVersion: simData.NodeTypeCuttingMachine},
// 		Name:                "Primary Cutter",
// 		FailureRate:         0.01,
// 		TimeSinceLastRepair: 0,
// 		Dullness:            0.0,
// 		Tools:               []string{"SteelBlade", "DiamondTip"},
// 	}, nil, 2*time.Second, queueSize)

// 	factory.AddNode("sensor1", &simData.SensorMachineNode{
// 		Node:          simData.Node{ID: "sensor1", NodeVersion: simData.NodeTypeSensorMachine},
// 		Name:          "Dimensional Scanner",
// 		Calibration:   1.0,
// 		FailureChance: 0.01,
// 	}, nil, 1*time.Second, queueSize)

// 	factory.AddEdges("start", "cutting1")
// 	factory.AddEdges("cutting1", "sensor1")
// 	factory.AddEdges("sensor1", "complete")

// 	if count := factory.GetCount(); count < 4 {
// 		t.Errorf("Nodes not added correctly, expected at least 4 but got: %d", count)
// 	}

// 	startNode := factory.GetNode("start")
// 	if startNode == nil {
// 		t.Fatal("Start node not found")
// 	}

// 	if len(startNode.GetNextNodes()) != 1 {
// 		t.Errorf("Start node should have 1 next node, but has %d", len(startNode.GetNextNodes()))
// 	}

// 	t.Log("Nodes added correctly")
// }

// func TestDataSourcesInitialization(t *testing.T) {
// 	connectors := GetConnector(t)

// 	table, data := createTest(t)

// 	err := connectors.AddData("postgres", table, data)
// 	if err != nil {
// 		t.Errorf("Error adding data: %v", err)
// 	}
// }

// func GetConnector(t *testing.T) connections.WorkspaceConnectors {
// 	regData := "workspaceConnectors"
// 	workspaceConnectorsObj, exists := reg.Get(regData)
// 	if !exists {
// 		t.Error("workspaceConnectors not found in registry")
// 	}
// 	if workspaceConnectorsObj == nil {
// 		t.Errorf("%v is nil", regData)
// 	}

// 	connectors := workspaceConnectorsObj.(connections.WorkspaceConnectors)

// 	return connectors
// }

// func createTest(t *testing.T) (connections.TableDefinition, []interface{}) {
// 	header := []connections.ColumnDefinition{
// 		{
// 			Name: "id",
// 			Type: connections.TypeInt,
// 		},
// 		{
// 			Name: "name",
// 			Type: connections.TypeText,
// 		},
// 		{
// 			Name: "example_value",
// 			Type: connections.TypeInt,
// 		},
// 	}

// 	table := connections.TableDefinition{
// 		Name:    "test_table",
// 		Schema:  "test",
// 		Columns: header,
// 	}
// 	rows := 20
// 	var data []interface{}
// 	data = createTestRows(t, table, rows)

// 	switch len(data) {
// 	case 0:
// 		t.Errorf("No data created")
// 	case rows:
// 	default:
// 		t.Errorf("Unexpected number of rows: got %d, expected %d", len(data), rows)
// 	}
// 	return table, data
// }

// func createTestRows(t *testing.T, table connections.TableDefinition, rows int) []interface{} {
// 	var data []interface{}

// 	header := table.Columns
// 	t.Logf("header: %v", header)
// 	for i := 0; i < rows; i++ {
// 		data = append(data, map[string]interface{}{
// 			header[0].Name: i,
// 			header[1].Name: "test",
// 			header[2].Name: 100,
// 		})
// 	}
// 	return data
// }
