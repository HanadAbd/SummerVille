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
	logging("%s - %s has entered the Factory\n", time.Now().Format("2006-01-02 3:4:5"), p.ID)
	for _, node := range n.NextNodes {
		return node
	}
	return n.ErrorNode
}

type Reject struct {
	Name string
}

func (r Reject) Process(n *Node, p *Part, connections map[string]*DataSources) *Node {
	logging(p.ID, " has been rejected")

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
	logging(p.ID, " has been completed")

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
	earliestNode := n.NextNodes["reject"]
	for _, node := range childNodes {
		node.NextNodes = n.NextNodes
		if node.Event == Idle {
			return node
		}
		if len(node.Queue) < len(earliestNode.Queue) {
			earliestNode = node
		}
	}
	if earliestNode != n.NextNodes["reject"] {
		return earliestNode
	}
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
	nodes map[string]*Node
}

func (f *Factory) AddNode(id string, nt NodeType, nw map[string]*Node, pt time.Duration) {
	nextNodes := make(map[string]*Node)
	errorNode := f.GetNode("reject")
	if errorNode == nil && id != "reject" {
		log.Fatal("Error node not found")
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

func (f *Factory) GetNode(id string) *Node {
	node := f.nodes[id]
	if node != nil {
		return node
	}
	return nil
}
func (f *Factory) GetAllNodes() map[string]*Node {
	return f.nodes
}

func (f *Factory) GetCount() int {
	return len(f.nodes)
}
func (f *Factory) UpdateNode(id string, nt NodeType, nw map[string]*Node, pt time.Duration) {
	node := f.GetNode(id)
	if node != nil {
		node.NodeType = nt
		node.NodesWithin = nw
		node.ProcessingTime = pt
	}
}
