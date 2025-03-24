package simData

import (
	"log"
	"time"
)

type Factory struct {
	nodes       map[string]*Node
	connections map[string]*DataSource
}

func (f *Factory) AddNode(id string, nt NodeType, nw map[string]*Node, pt time.Duration, queueSize int) {
	if queueSize <= 0 {
		queueSize = 100
	}
	nextNodes := make(map[string]*Node)
	errorNode := f.GetNode("reject")
	if errorNode == nil && id != "reject" {
		log.Println("Error node not found")
	}

	f.nodes[id] = &Node{
		ID:             id,
		NodeType:       nt,
		NodesWithin:    nw,
		NextNodes:      nextNodes,
		Queue:          make(chan *Part, queueSize),
		Event:          Idle,
		ProcessingTime: pt,
		ErrorNode:      errorNode,
	}
}

func (factory *Factory) AddEdges(from string, to string) {
	start := factory.GetNode(from)
	end := factory.GetNode(to)
	start.NextNodes[to] = end
}

func (f *Factory) UpdateNode(id string, nt NodeType, nw map[string]*Node, pt time.Duration) {
	node := f.GetNode(id)
	if node != nil {
		node.NodeType = nt
		node.NodesWithin = nw
		node.ProcessingTime = pt
	}
}

func (f *Factory) GetNode(id string) *Node {
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
		nodes[node.ID] = map[string]interface{}{
			"id":             node.ID,
			"nodeType":       node.NodeType.GetName(),
			"nodesWithin":    allKeys(node.NodesWithin),
			"nextNodes":      allKeys(node.NextNodes),
			"queue":          len(node.Queue),
			"event":          node.Event.String(),
			"processingTime": node.ProcessingTime.Seconds(),
		}
	}
	return nodes
}

func allKeys(m map[string]*Node) []string {
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
			"id":             node.ID,
			"nodesWithin":    allKeys(node.NodesWithin),
			"nextNodes":      allKeys(node.NextNodes),
			"queue":          len(node.Queue),
			"event":          node.Event.String(),
			"processingTime": node.ProcessingTime.Seconds(),
		}
	}
	return nil
}
func (f *Factory) SetNodeData(id string, node *Node) {
	f.nodes[id].ProcessingTime = node.ProcessingTime
	f.nodes[id].NodesWithin = node.NodesWithin
	f.nodes[id].NextNodes = node.NextNodes

}

func IntiliaseFactory(connections map[string]*DataSource) *Factory {
	factory := &Factory{
		nodes:       make(map[string]*Node),
		connections: connections,
	}
	queueSize := 100

	factory.AddNode("reject", Reject{Name: "reject"}, nil, 0, queueSize)
	factory.AddNode("start", Start{Name: "start"}, nil, 0, queueSize)
	factory.AddNode("complete", Complete{Name: "complete"}, nil, 0, queueSize)

	factory.AddNode("cutting1", CuttingMachine{Name: "cutting1"}, nil, 2*time.Second, queueSize)
	factory.AddNode("cutting2", CuttingMachine{Name: "cutting2"}, nil, 2*time.Second, queueSize)
	factory.AddNode("cutting3", CuttingMachine{Name: "cutting3"}, nil, 2*time.Second, queueSize)

	factory.AddNode("sensor1", Sensor{Name: "sensor1"}, nil, time.Second, queueSize)
	factory.AddNode("sensor2", Sensor{Name: "sensor2"}, nil, time.Second, queueSize)
	factory.AddNode("sensor3", Sensor{Name: "sensor3"}, nil, time.Second, queueSize)

	factory.AddNode("station1", Station{Name: "station1"}, stationMap(factory, "cutting1", "cutting2", "cutting3"), 0, queueSize)
	factory.AddNode("station2", Station{Name: "station2"}, stationMap(factory, "sensor1", "sensor2", "sensor3"), 0, queueSize)

	factory.AddEdges("start", "station1")
	factory.AddEdges("station1", "station2")
	factory.AddEdges("station2", "complete")
	factory.AddEdges("station2", "reject")

	return factory
}
func stationMap(factory *Factory, names ...string) map[string]*Node {
	nodes := make(map[string]*Node)
	for _, name := range names {
		nodes[name] = factory.GetNode(name)
	}
	return nodes
}
