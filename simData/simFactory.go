package simData

import (
	"log"
	"time"
)

type Factory struct {
	nodes       map[string]FactoryNode
	connections map[string]*DataSource
}

func (f *Factory) AddNode(id string, node FactoryNode, nodesWithin map[string]FactoryNode, processingTime time.Duration, queueSize int) {
	if queueSize <= 0 {
		queueSize = 500
	}
	nextNodes := make(map[string]FactoryNode)
	errorNode := f.GetNode("reject")
	if errorNode == nil && id != "reject" {
		log.Println("Error node not found")
	}

	f.nodes[id] = node
	f.nodes[id].SetID(id)
	f.nodes[id].SetType(node.Type())
	f.nodes[id].SetNodesWithin(nodesWithin)
	f.nodes[id].SetNextNodes(nextNodes)
	f.nodes[id].SetQueue(make(chan *Part, queueSize))
	f.nodes[id].SetEvent(Idle)
	f.nodes[id].SetProcessingTime(processingTime)
	f.nodes[id].SetErrorNode(errorNode)

}

func (factory *Factory) AddEdges(from string, to string) {
	start := factory.GetNode(from)
	end := factory.GetNode(to)

	if start == nil || end == nil {
		return
	}

	start.GetNextNodes()[end.GetID()] = end
}

func (f *Factory) GetNode(id string) FactoryNode {
	node := f.nodes[id]
	if node != nil {
		return node
	}
	return nil
}

func (f *Factory) GetCount() int {
	return len(f.nodes)
}

func (f *Factory) GetAllNodes() map[string]interface{} {
	nodes := make(map[string]interface{})
	for _, node := range f.nodes {
		nodes[node.GetID()] = map[string]interface{}{
			"id":             node.GetID(),
			"nodeType":       node.GetType(),
			"nodesWithin":    allKeys(node.GetNodesWithin()),
			"nextNodes":      allKeys(node.GetNextNodes()),
			"queue":          len(node.GetQueue()),
			"event":          node.GetEvent().String(),
			"processingTime": node.GetProcessingTime().Seconds(),
		}
	}
	return nodes
}

func allKeys(m map[string]FactoryNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (f *Factory) GetAllReports() map[string]interface{} {
	reports := make(map[string]interface{})
	for _, conn := range f.connections {
		reports[conn.Name] = map[string]interface{}{
			"name": conn.Name,
			"type": conn.DataType,
			// TODO add these fields
			// "status":     conn.Status,
			// "frequency":  conn.Frequency,
			// "lastUpdate": conn.LastUpdate,
			// "dataAdded":  conn.DataAdded,
		}
	}
	return reports
}

func (f *Factory) GetNodeData(id string) map[string]interface{} {
	node := f.GetNode(id)
	if node != nil {
		return map[string]interface{}{
			"node_id":         node.GetID(),
			"nodes_within":    allKeys(node.GetNodesWithin()),
			"next_nodes":      allKeys(node.GetNextNodes()),
			"queue":           len(node.GetQueue()),
			"node_event":      node.GetEvent().String(),
			"processing_time": node.GetProcessingTime().Seconds(),
		}
	}
	return nil
}
func (f *Factory) SetNodeData(id string, node FactoryNode) {
	f.nodes[id].SetProcessingTime(node.GetProcessingTime())
	f.nodes[id].SetNodesWithin(node.GetNodesWithin())
	f.nodes[id].SetNextNodes(node.GetNextNodes())

}

func IntiliaseFactory(connections map[string]*DataSource) *Factory {
	factory := &Factory{
		nodes:       make(map[string]FactoryNode),
		connections: connections,
	}
	queueSize := 500

	// Base nodes
	factory.AddNode("reject", &Reject{Name: "Reject Bin"}, nil, 0, queueSize)
	factory.AddNode("start", &Start{Name: "Production Start"}, nil, 0, queueSize)
	factory.AddNode("complete", &Complete{Name: "Production Complete"}, nil, 0, queueSize)

	// Inventory nodes
	factory.AddNode("raw_inventory", &InventoryNode{
		Node:          Node{ID: "raw_inventory", NodeVersion: NodeTypeInventory},
		Name:          "Raw Materials Storage",
		Capacity:      500,
		AllowedTypes:  []string{"Steel", "Aluminum", "Plastic", "Electronics"},
		CurrentStored: 0,
	}, nil, 1*time.Second, queueSize)

	factory.AddNode("component_inventory", &InventoryNode{
		Node:          Node{ID: "component_inventory", NodeVersion: NodeTypeInventory},
		Name:          "Component Storage",
		Capacity:      150,
		AllowedTypes:  []string{"Steel", "Aluminum", "Plastic", "Electronics", "Mixed"},
		CurrentStored: 0,
	}, nil, 1*time.Second, queueSize)

	factory.AddNode("finished_inventory", &InventoryNode{
		Node:          Node{ID: "finished_inventory", NodeVersion: NodeTypeInventory},
		Name:          "Finished Goods Storage",
		Capacity:      500,
		AllowedTypes:  []string{"Steel", "Aluminum", "Plastic", "Electronics", "Mixed"},
		CurrentStored: 0,
	}, nil, 1*time.Second, queueSize)

	// Cutting Department
	factory.AddNode("cutting1", &CuttingMachineNode{
		Node:                Node{ID: "cutting1", NodeVersion: NodeTypeCuttingMachine},
		Name:                "Primary Cutter",
		FailureRate:         0.01,
		TimeSinceLastRepair: 0,
		Dullness:            0.0,
		Tools:               []string{"SteelBlade", "DiamondTip"},
	}, nil, 2*time.Second, queueSize)

	factory.AddNode("cutting2", &CuttingMachineNode{
		Node:                Node{ID: "cutting2", NodeVersion: NodeTypeCuttingMachine},
		Name:                "Secondary Cutter",
		FailureRate:         0.02,
		TimeSinceLastRepair: 0,
		Dullness:            0.0,
		Tools:               []string{"TitaniumBlade", "CarbideTip"},
	}, nil, 2*time.Second, queueSize)

	factory.AddNode("cutting3", &CuttingMachineNode{
		Node:                Node{ID: "cutting3", NodeVersion: NodeTypeCuttingMachine},
		Name:                "Precision Cutter",
		FailureRate:         0.005,
		TimeSinceLastRepair: 0,
		Dullness:            0.0,
		Tools:               []string{"LaserCutter", "WaterJet"},
	}, nil, 3*time.Second, queueSize)

	factory.AddNode("cutting_worker", &WorkerNode{
		Node:       Node{ID: "cutting_worker", NodeVersion: NodeTypeWorker},
		Name:       "John Smith",
		Department: "Cutting Department",
		SkillLevel: 4,
	}, nil, 2*time.Second, queueSize)

	// Quality Control Department
	factory.AddNode("sensor1", &SensorMachineNode{
		Node:          Node{ID: "sensor1", NodeVersion: NodeTypeSensorMachine},
		Name:          "Dimensional Scanner",
		Calibration:   1.0,
		FailureChance: 0.01,
	}, nil, 1*time.Second, queueSize)

	factory.AddNode("sensor2", &SensorMachineNode{
		Node:          Node{ID: "sensor2", NodeVersion: NodeTypeSensorMachine},
		Name:          "Surface Analyzer",
		Calibration:   0.98,
		FailureChance: 0.02,
	}, nil, 1*time.Second, queueSize)

	factory.AddNode("sensor3", &SensorMachineNode{
		Node:          Node{ID: "sensor3", NodeVersion: NodeTypeSensorMachine},
		Name:          "Weight Verifier",
		Calibration:   1.02,
		FailureChance: 0.015,
	}, nil, 1*time.Second, queueSize)

	factory.AddNode("qc_worker", &WorkerNode{
		Node:       Node{ID: "qc_worker", NodeVersion: NodeTypeWorker},
		Name:       "Alice Johnson",
		Department: "Quality Control",
		SkillLevel: 5,
	}, nil, 2*time.Second, queueSize)

	// Repair Department
	factory.AddNode("repair1", &RepairStationNode{
		Node:           Node{ID: "repair1", NodeVersion: NodeTypeRepairStation},
		Name:           "Minor Defect Repair",
		RepairCapacity: 1,
	}, nil, 4*time.Second, queueSize)

	factory.AddNode("repair2", &RepairStationNode{
		Node:           Node{ID: "repair2", NodeVersion: NodeTypeRepairStation},
		Name:           "Major Defect Repair",
		RepairCapacity: 3,
	}, nil, 6*time.Second, queueSize)

	factory.AddNode("repair_worker", &WorkerNode{
		Node:       Node{ID: "repair_worker", NodeVersion: NodeTypeWorker},
		Name:       "Robert Chen",
		Department: "Repair Department",
		SkillLevel: 4,
	}, nil, 3*time.Second, queueSize)

	// Assembly Department
	factory.AddNode("assembly1", &AssemblyStationNode{
		Node:          Node{ID: "assembly1", NodeVersion: NodeTypeAssemblyStation},
		Name:          "Component Assembly",
		ToolsRequired: []string{"Screwdriver", "Pliers", "Hammer"},
	}, nil, 5*time.Second, queueSize)

	factory.AddNode("assembly2", &AssemblyStationNode{
		Node:          Node{ID: "assembly2", NodeVersion: NodeTypeAssemblyStation},
		Name:          "Electronics Assembly",
		ToolsRequired: []string{"Soldering Iron", "Multimeter", "Tweezers"},
	}, nil, 7*time.Second, queueSize)

	factory.AddNode("assembly_worker1", &WorkerNode{
		Node:       Node{ID: "assembly_worker1", NodeVersion: NodeTypeWorker},
		Name:       "Lisa Wang",
		Department: "Assembly",
		SkillLevel: 5,
	}, nil, 2*time.Second, queueSize)

	factory.AddNode("assembly_worker2", &WorkerNode{
		Node:       Node{ID: "assembly_worker2", NodeVersion: NodeTypeWorker},
		Name:       "David Martin",
		Department: "Assembly",
		SkillLevel: 4,
	}, nil, 2*time.Second, queueSize)

	// Packaging Department
	factory.AddNode("packaging1", &PackagingNode{
		Node:          Node{ID: "packaging1", NodeVersion: NodeTypePackaging},
		Name:          "Box Packaging",
		PackagingType: "Cardboard Box",
	}, nil, 2*time.Second, queueSize)

	factory.AddNode("packaging2", &PackagingNode{
		Node:          Node{ID: "packaging2", NodeVersion: NodeTypePackaging},
		Name:          "Premium Packaging",
		PackagingType: "Clamshell",
	}, nil, 3*time.Second, queueSize)

	factory.AddNode("packaging_worker", &WorkerNode{
		Node:       Node{ID: "packaging_worker", NodeVersion: NodeTypeWorker},
		Name:       "Bob Williams",
		Department: "Packaging",
		SkillLevel: 3,
	}, nil, 2*time.Second, queueSize)

	// Final Inspection
	factory.AddNode("final_inspection", &SensorMachineNode{
		Node:          Node{ID: "final_inspection", NodeVersion: NodeTypeSensorMachine},
		Name:          "Final Product Inspection",
		Calibration:   1.0,
		FailureChance: 0.005,
	}, nil, 4*time.Second, queueSize)

	factory.AddNode("inspection_worker", &WorkerNode{
		Node:       Node{ID: "inspection_worker", NodeVersion: NodeTypeWorker},
		Name:       "Sarah Johnson",
		Department: "Final Inspection",
		SkillLevel: 5,
	}, nil, 3*time.Second, queueSize)

	// Production line departments (grouping nodes)
	factory.AddNode("cutting_station", &Station{
		Name: "Cutting Department",
	}, stationMap(factory, "cutting1", "cutting2", "cutting3", "cutting_worker"), 0, queueSize)

	factory.AddNode("qc_station", &Station{
		Name: "Quality Control",
	}, stationMap(factory, "sensor1", "sensor2", "sensor3", "qc_worker"), 0, queueSize)

	factory.AddNode("repair_station", &Station{
		Name: "Repair Department",
	}, stationMap(factory, "repair1", "repair2", "repair_worker"), 0, queueSize)

	factory.AddNode("assembly_station", &Station{
		Name: "Assembly Department",
	}, stationMap(factory, "assembly1", "assembly2", "assembly_worker1", "assembly_worker2"), 0, queueSize)

	factory.AddNode("packaging_station", &Station{
		Name: "Packaging Department",
	}, stationMap(factory, "packaging1", "packaging2", "packaging_worker"), 0, queueSize)

	factory.AddNode("inspection_station", &Station{
		Name: "Final Inspection",
	}, stationMap(factory, "final_inspection", "inspection_worker"), 0, queueSize)

	// Define the complete production flow
	// Main flow: start -> raw_inventory -> cutting_station -> qc_station -> component_inventory ->
	//   assembly_station -> qc_station -> packaging_station -> inspection_station -> finished_inventory -> complete
	factory.AddEdges("start", "raw_inventory")
	factory.AddEdges("raw_inventory", "cutting_station")
	factory.AddEdges("cutting_station", "qc_station")
	factory.AddEdges("qc_station", "component_inventory")
	factory.AddEdges("component_inventory", "assembly_station")
	factory.AddEdges("assembly_station", "qc_station")
	factory.AddEdges("qc_station", "packaging_station")
	factory.AddEdges("packaging_station", "inspection_station")
	factory.AddEdges("inspection_station", "finished_inventory")
	factory.AddEdges("finished_inventory", "complete")

	// Error handling and alternative paths
	factory.AddEdges("cutting_station", "repair_station")
	factory.AddEdges("qc_station", "repair_station")
	factory.AddEdges("repair_station", "qc_station")
	factory.AddEdges("inspection_station", "repair_station")

	// Reject paths
	factory.AddEdges("repair_station", "reject")
	factory.AddEdges("qc_station", "reject")
	factory.AddEdges("inspection_station", "reject")

	// Worker direct connections
	factory.AddEdges("cutting_worker", "cutting1")
	factory.AddEdges("cutting_worker", "cutting2")
	factory.AddEdges("cutting_worker", "cutting3")

	factory.AddEdges("qc_worker", "sensor1")
	factory.AddEdges("qc_worker", "sensor2")
	factory.AddEdges("qc_worker", "sensor3")

	factory.AddEdges("repair_worker", "repair1")
	factory.AddEdges("repair_worker", "repair2")

	factory.AddEdges("assembly_worker1", "assembly1")
	factory.AddEdges("assembly_worker2", "assembly2")

	factory.AddEdges("packaging_worker", "packaging1")
	factory.AddEdges("packaging_worker", "packaging2")

	factory.AddEdges("inspection_worker", "final_inspection")

	// Special routing cases
	factory.AddEdges("cutting3", "sensor1")  // Precision cuts always go to dimensional scanner
	factory.AddEdges("assembly2", "sensor2") // Electronics assembly always gets surface analysis

	// Allow direct transfers between departments when needed
	factory.AddEdges("assembly_station", "packaging_station")    // Fast track for simple products
	factory.AddEdges("component_inventory", "packaging_station") // Pre-assembled components

	return factory
}
func stationMap(factory *Factory, names ...string) map[string]FactoryNode {
	nodes := make(map[string]FactoryNode)
	for _, name := range names {
		nodes[name] = factory.GetNode(name)
	}
	return nodes
}
