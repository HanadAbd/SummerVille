package simData

import (
	"log"
	"math/rand"
	"time"
)

type Start struct {
	Name string
}

func (s Start) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	logPartState(p.ID, "Started", n.ID)

	for _, node := range n.NextNodes {
		return node
	}
	return n.ErrorNode
}

type Reject struct {
	Name string
}

func (r Reject) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	logPartState(p.ID, "Rejected", n.ID)

	condition := logCondition(n, p, func(n *Node, p *Part) (bool, error) {
		return true, nil
	})
	data := &Data{Header: []string{"part_id", "cut_attempts", "cut_val", "rejected_at"}, rows: [][]interface{}{{p.ID, p.Cutattempts, p.CutVal, time.Now()}}}
	connections["reject"].Appender(condition, connections["reject"], data)
	return nil
}

type Complete struct {
	Name string
}

func (c Complete) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	logPartState(p.ID, "Completed", n.ID)

	condition := logCondition(n, p, func(n *Node, p *Part) (bool, error) {
		return true, nil
	})
	data := &Data{Header: []string{"part_id", "cut_attempts", "cut_val", "rejected_at"}, rows: [][]interface{}{{p.ID, p.Cutattempts, p.CutVal, time.Now()}}}
	connections["cutting"].Appender(condition, connections["cutting"], data)

	return nil
}

type Station struct {
	Name string
}

func (s Station) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	childNodes := n.NodesWithin

	// Check if reject node exists before accessing it
	var rejectNode *Node
	var exists bool
	if rejectNode, exists = n.NextNodes["reject"]; !exists || rejectNode == nil {
		// Fallback to error node if reject node doesn't exist
		rejectNode = n.ErrorNode
		// If error node is also nil, just use the first available next node
		if rejectNode == nil && len(n.NextNodes) > 0 {
			for _, node := range n.NextNodes {
				rejectNode = node
				break
			}
		}
		// If still nil, log warning and return nil (which will end processing)
		if rejectNode == nil {
			logging("Warning: No reject node found for station %s and no fallback available\n", n.ID)
			return nil
		}
	}

	earliestNode := rejectNode

	for _, node := range childNodes {
		// Make sure all nodes have proper next node references
		if len(node.NextNodes) == 0 {
			node.NextNodes = n.NextNodes
		}

		// Return the first idle node immediately
		if node.Event == Idle {
			return node
		}

		// Find the node with the shortest queue
		if earliestNode == nil || len(node.Queue) < len(earliestNode.Queue) {
			earliestNode = node
		}
	}

	// If we found a node with queue shorter than reject's queue, use it
	if earliestNode != nil && earliestNode != rejectNode {
		return earliestNode
	}

	// Otherwise use the error node
	return n.ErrorNode
}

type CuttingMachine struct {
	Name    string
	Station *Node
}

func (c CuttingMachine) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	p.CutVal = rand.Intn(16)
	p.Cutattempts += 1

	for _, node := range n.NextNodes {
		return node
	}
	return n.ErrorNode
}

type Sensor struct {
	Name    string
	Station *Node
}

func (s Sensor) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	minAcceptable, maxAcceptable := 8, 12
	if p.Cutattempts >= 2 {
		return n.NextNodes["reject"]
	}
	if p.CutVal < minAcceptable || p.CutVal > maxAcceptable {
		if p.CutVal < 4 || p.CutVal > 14 {
			return n.NextNodes["reject"]
		}
		return n.NextNodes["station1"]
	}

	condition := logCondition(n, p, func(n *Node, p *Part) (bool, error) {
		return true, nil
	})
	data := &Data{Header: []string{"part_id", "cut_attempts", "cut_val", "rejected_at"}, rows: [][]interface{}{{p.ID, p.Cutattempts, p.CutVal, time.Now()}}}
	connections["sensor"].Appender(condition, connections["sensor"], data)
	return n.NextNodes["complete"]
}

type Factory struct {
	nodes       map[string]*Node
	connections map[string]*DataSources
}

func (f *Factory) AddNode(id string, nt NodeType, nw map[string]*Node, pt time.Duration) {
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
		Queue:          make(chan *Part, 100),
		Event:          Idle,
		ProcessingTime: pt,
		ErrorNode:      errorNode,
	}
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
			"name":       conn.Name,
			"type":       conn.DataType,
			"frequency":  conn.Frequency,
			"lastUpdate": conn.LastUpdate,
			"dataAdded":  conn.DataAdded,
		}
	}
	return reports
}
