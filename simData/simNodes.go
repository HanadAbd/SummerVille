package simData

import (
	"context"
	"foo/backend/connections"
	"foo/services/util"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type MachineState int

const (
	Idle MachineState = iota
	Processing
	Processed
	Faulty
)

func (s MachineState) String() string {
	switch s {
	case Idle:
		return "Idle"
	case Processing:
		return "Processing"
	case Processed:
		return "Processed"
	case Faulty:
		return "Faulty"
	}
	return "Unknown"
}

type Part struct {
	ID          string
	NodeHistory []*Node
	Cutattempts int
	CutVal      int
}

type NodeType interface {
	Process(n *Node, p *Part, c map[string]*DataSource) *Node
	GetName() string
}

type Node struct {
	ID             string
	NodeType       NodeType
	NodesWithin    map[string]*Node
	NextNodes      map[string]*Node
	Queue          chan *Part
	Event          MachineState
	ProcessingTime time.Duration
	ErrorNode      *Node
}

func (n *Node) Start(wg *sync.WaitGroup, connections map[string]*DataSource, ctx context.Context) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, exiting node %s", n.ID)
			return

		case part, ok := <-n.Queue:
			if !ok {
				log.Printf("Queue closed, exiting node %s", n.ID)
				return
			}

			// Process the part
			logPartState(part.ID, "Processing", n.ID)
			n.Event = Processing

			time.Sleep(n.ProcessingTime)
			part.NodeHistory = append(part.NodeHistory, n)

			logPartState(part.ID, "Processed", n.ID)
			n.Event = Processed

			nextNode := n.NodeType.Process(n, part, connections)

			if nextNode != nil {
				logPartTransition(part.ID, n.ID, nextNode.ID)
				time.Sleep(time.Second)

				select {
				case nextNode.Queue <- part:

				case <-time.After(500 * time.Millisecond):
					log.Printf("Next node queue full, dropping part %s", part.ID)
				case <-ctx.Done():
					log.Printf("Context cancelled while sending to next node, exiting %s", n.ID)
					return
				}

				n.Event = Idle
				logPartState(part.ID, "Idle", n.ID)

				queueContents := getQueueContents(n.Queue)
				if len(queueContents) > 0 {
					logNodeQueue(n.ID, queueContents)
				}
			}

		default:
			// Check for idle state
			if n.Event != Idle {
				logPartState("", "Idle", n.ID)
				n.Event = Idle
			}

			// Don't busy-wait, sleep a bit
			select {
			case <-time.After(100 * time.Millisecond):
				// Just a delay to prevent CPU spinning
			case <-ctx.Done():
				log.Printf("Context cancelled during idle, exiting node %s", n.ID)
				return
			}
		}
	}
}

type Start struct {
	Name string
}

func (s Start) Process(n *Node, p *Part, connections map[string]*DataSource) *Node {
	logPartState(p.ID, "Started", n.ID)

	for _, node := range n.NextNodes {
		return node
	}
	return n.ErrorNode
}

func (s Start) GetName() string {
	return "Start"
}

type Reject struct {
	Name string
}

func (r Reject) Process(n *Node, p *Part, connections map[string]*DataSource) *Node {
	logPartState(p.ID, "Rejected", n.ID)

	if conn, exists := connections["reject"]; exists {
		conn.Appender(p, n, conn.DataMapper(p, n))
	}

	return nil
}

func (r Reject) GetName() string {
	return "Complete"
}

type Complete struct {
	Name string
}

func (c Complete) Process(n *Node, p *Part, connections map[string]*DataSource) *Node {
	logPartState(p.ID, "Completed", n.ID)

	if conn, exists := connections["complete"]; exists {
		conn.Appender(p, n, conn.DataMapper(p, n))
	}

	return nil
}

func (c Complete) GetName() string {
	return "Complete"
}

type Station struct {
	Name string
}

func (s Station) Process(n *Node, p *Part, connections map[string]*DataSource) *Node {
	childNodes := n.NodesWithin

	var rejectNode *Node
	var exists bool
	if rejectNode, exists = n.NextNodes["reject"]; !exists || rejectNode == nil {
		rejectNode = n.ErrorNode
		if rejectNode == nil && len(n.NextNodes) > 0 {
			for _, node := range n.NextNodes {
				rejectNode = node
				break
			}
		}
		if rejectNode == nil {
			logging("Warning: No reject node found for station %s and no fallback available\n", n.ID)
			return nil
		}
	}

	earliestNode := rejectNode

	for _, node := range childNodes {
		if len(node.NextNodes) == 0 {
			node.NextNodes = n.NextNodes
		}

		if node.Event == Idle {
			return node
		}

		if earliestNode == nil || len(node.Queue) < len(earliestNode.Queue) {
			earliestNode = node
		}
	}

	if earliestNode != nil && earliestNode != rejectNode {
		return earliestNode
	}

	return n.ErrorNode
}

func (c Station) GetName() string {
	return "Station"
}

type CuttingMachine struct {
	Name    string
	Station *Node
}

func (c CuttingMachine) Process(n *Node, p *Part, connections map[string]*DataSource) *Node {
	p.CutVal = rand.Intn(16)
	p.Cutattempts += 1

	for _, node := range n.NextNodes {
		return node
	}
	return n.ErrorNode
}

func (c CuttingMachine) GetName() string {
	return "Cutting Machine"
}

type Sensor struct {
	Name    string
	Station *Node
}

func (s Sensor) Process(n *Node, p *Part, connections map[string]*DataSource) *Node {
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

	// Add data to the sensor data source
	if conn, exists := connections["sensor"]; exists {
		conn.Appender(p, n, conn.DataMapper(p, n))
	}

	// Also record throughput metrics
	if conn, exists := connections["station_metrics"]; exists {
		conn.Appender(p, n, conn.DataMapper(p, n))
	}

	return n.NextNodes["complete"]
}

func (s Sensor) GetName() string {
	return "Sensor"
}

func IntialiseConnections(registry *util.Registry) map[string]*DataSource {
	conns := make(map[string]*DataSource)

	// Cutting machine data source
	conns["cutting"] = &DataSource{
		Name:     "cutting",
		DataType: "kafka",
		Table: &connections.TableDefinition{
			Name:   "cutting",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "cut_attempts", Type: connections.TypeInt, Nullable: false},
				{Name: "cut_val", Type: connections.TypeInt, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return p.Cutattempts > 0
		},
		DataMapper: func(p *Part, n *Node) map[string]interface{} {
			return map[string]interface{}{
				"part_id":      p.ID,
				"cut_attempts": p.Cutattempts,
				"cut_val":      p.CutVal,
				"timestamp":    time.Now(),
			}
		},
	}

	// Sensor data source (quality checks)
	conns["sensor"] = &DataSource{
		Name:     "sensor",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "quality_checks",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "quality_value", Type: connections.TypeInt, Nullable: false},
				{Name: "is_acceptable", Type: connections.TypeBoolean, Nullable: false},
				{Name: "sensor_id", Type: connections.TypeText, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return strings.HasPrefix(n.ID, "sensor")
		},
		DataMapper: func(p *Part, n *Node) map[string]interface{} {
			minAcceptable, maxAcceptable := 8, 12
			isAcceptable := p.CutVal >= minAcceptable && p.CutVal <= maxAcceptable

			return map[string]interface{}{
				"part_id":       p.ID,
				"quality_value": p.CutVal,
				"is_acceptable": isAcceptable,
				"sensor_id":     n.ID,
				"timestamp":     time.Now(),
			}
		},
	}

	// Add station metrics data source
	conns["station_metrics"] = &DataSource{
		Name:     "station_metrics",
		DataType: "csv",
		Table: &connections.TableDefinition{
			Name:   "station_throughput",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "station_id", Type: connections.TypeText, Nullable: false},
				{Name: "queue_size", Type: connections.TypeInt, Nullable: false},
				{Name: "station_type", Type: connections.TypeText, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return strings.HasPrefix(n.ID, "station")
		},
		DataMapper: func(p *Part, n *Node) map[string]interface{} {
			return map[string]interface{}{
				"station_id":   n.ID,
				"queue_size":   len(n.Queue),
				"station_type": n.NodeType.GetName(),
				"timestamp":    time.Now(),
			}
		},
	}

	// Add reject and complete processing data sources
	conns["reject"] = &DataSource{
		Name:     "reject",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "rejected_parts",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "cut_val", Type: connections.TypeInt, Nullable: false},
				{Name: "cut_attempts", Type: connections.TypeInt, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return n.ID == "reject"
		},
		DataMapper: func(p *Part, n *Node) map[string]interface{} {
			return map[string]interface{}{
				"part_id":      p.ID,
				"cut_val":      p.CutVal,
				"cut_attempts": p.Cutattempts,
				"timestamp":    time.Now(),
			}
		},
	}

	conns["complete"] = &DataSource{
		Name:     "complete",
		DataType: "kafka",
		Table: &connections.TableDefinition{
			Name:   "completed_parts",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "processing_time", Type: connections.TypeFloat, Nullable: false},
				{Name: "quality_score", Type: connections.TypeInt, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return n.ID == "complete"
		},
		DataMapper: func(p *Part, n *Node) map[string]interface{} {
			qualityScore := 10 - abs(10-p.CutVal)
			if qualityScore < 0 {
				qualityScore = 0
			}

			return map[string]interface{}{
				"part_id":         p.ID,
				"processing_time": float64(len(p.NodeHistory)) * 1.5,
				"quality_score":   qualityScore,
				"timestamp":       time.Now(),
			}
		},
	}

	log.Println("Data sources initialized")
	return conns
}
