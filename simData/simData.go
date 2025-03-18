package simData

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"foo/services/util"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
)

type DataType int

const (
	CSV DataType = iota
	Postgres
	Kafka
)

func (d DataType) String() string {
	switch d {
	case CSV:
		return "CSV"
	case Postgres:
		return "Postgres"
	case Kafka:
		return "Kafka"
	}
	return "Unknown"
}

type Data struct {
	Header []string
	rows   [][]interface{}
}

type DataSources struct {
	Name              string
	DataType          DataType
	ConnectionDetails interface{}
	ConnectionPoint   interface{}
	Data              *Data
	NodeID            []string
	Frequency         int
	LastUpdate        time.Time
	DataAdded         float64
}

func (d *DataSources) Appender(condition bool, c *DataSources, data *Data) {
	if !condition {
		return
	}
	if c.Data == nil {
		intialiseSource(c, data.Header)
		logging("Initialised %s\n", c.Name)
	}
	c.Data.rows = append(c.Data.rows, data.rows[0])
	addToSource(c, data.rows)
	c.Data.rows = nil
}
func logCondition(n *Node, p *Part, condition func(n *Node, p *Part) (bool, error)) bool {
	result, err := condition(n, p)
	if err != nil {
		logging("Error in condition: %v\n", err)
		return false
	}
	return result
}

func addToSource(c *DataSources, row [][]interface{}) {
	switch c.DataType {
	case CSV:
		addData(HandleCSV, c, row)
	case Postgres:
		addData(HandlePostgres, c, row)
	case Kafka:
		addData(HandleKafka, c, row)
	}
}
func addData(handler func(*DataSources, [][]interface{}) error, c *DataSources, row [][]interface{}) {
	if err := handler(c, row); err != nil {
		logging("Error adding data to %s: %v\n", c.DataType, err)
		return
	}
	logging("Added data to %s\n", c.Name)
}

var logPath = ""
var reg *util.Registry

func SetRegistry(registry *util.Registry) {
	reg = registry
}

func logging(format string, a ...any) (n int, err error) {
	message := []byte(fmt.Sprintf(format, a...))

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	n, err = file.Write(message)
	if err != nil {
		return 0, err
	}
	if reg != nil {
		reg.BroadcastToChannel("logs", message)
	}
	return n, nil
}

// New helper functions for structured logging
func logPartState(partID string, state string, nodeID string) {
	logMessage := fmt.Sprintf("%s;state;%s;%s\n", partID, state, nodeID)
	logging(logMessage)
}

func logPartTransition(partID string, sourceNodeID string, targetNodeID string) {
	logMessage := fmt.Sprintf("%s;transition;%s;%s\n", partID, sourceNodeID, targetNodeID)
	logging(logMessage)
}

func logNodeQueue(nodeID string, queueContents []string) {
	logMessage := fmt.Sprintf("%s;queue;%s\n", nodeID, strings.Join(queueContents, ","))
	logging(logMessage)
}

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

type NodeType interface {
	Process(n *Node, p *Part, c map[string]*DataSources) *Node
}

func (n *Node) Start(wg *sync.WaitGroup, connections map[string]*DataSources) {
	defer wg.Done()

	for part := range n.Queue {
		// Log part entering node and state change
		logPartState(part.ID, "Processing", n.ID)
		n.Event = Processing

		// Process the part
		time.Sleep(n.ProcessingTime)
		part.NodeHistory = append(part.NodeHistory, n)

		// Mark as processed
		logPartState(part.ID, "Processed", n.ID)
		n.Event = Processed

		// Find next node
		nextNode := n.NodeType.Process(n, part, connections)

		if nextNode != nil {
			// Log transition start (takes 1 second)
			logPartTransition(part.ID, n.ID, nextNode.ID)
			time.Sleep(time.Second) // Add 1 second transition time

			// Send to next node
			nextNode.Queue <- part

			// Reset to idle after processing
			n.Event = Idle
			logPartState(part.ID, "Idle", n.ID)

			// Log queue contents after processing
			queueContents := getQueueContents(n.Queue)
			if len(queueContents) > 0 {
				logNodeQueue(n.ID, queueContents)
			}
		}
	}
}
func getQueueContents(queue chan *Part) []string {
	// This is a non-blocking way to check queue contents
	// without consuming the queue
	contents := []string{}

	// Create a snapshot of the queue (non-blocking)
	queueLen := len(queue)
	if queueLen == 0 {
		return contents
	}

	// Try to get a snapshot of queue contents without blocking
	tempQueue := make([]string, 0, queueLen)

	for i := 0; i < queueLen; i++ {
		select {
		case part := <-queue:
			tempQueue = append(tempQueue, part.ID)
			queue <- part // Put it back
		default:
			// Queue changed while we were reading it
			// Just return what we have so far
			return tempQueue
		}
	}

	return tempQueue
}

func addEdges(factory *Factory, from string, to string) {
	start := factory.GetNode(from)
	end := factory.GetNode(to)
	start.NextNodes[to] = end
}
func addParts(start *Node, rate int, wg *sync.WaitGroup) {
	defer wg.Done()
	counter := 0

	for {
		counter++
		part := &Part{ID: "part" + fmt.Sprint(counter), Cutattempts: 0}
		start.Queue <- part
		duration := float64(time.Second) / ((rand.Float64() * float64(rate) / 2) + 0.75)
		time.Sleep(time.Duration(duration))
	}
}

func stationMap(factory *Factory, names ...string) map[string]*Node {
	nodes := make(map[string]*Node)
	for _, name := range names {
		nodes[name] = factory.GetNode(name)
	}
	return nodes
}

func IntiliaseFactory(connections map[string]*DataSources) *Factory {
	factory := &Factory{
		nodes:       make(map[string]*Node),
		connections: connections,
	}
	factory.AddNode("reject", Reject{Name: "reject"}, nil, 0)
	factory.AddNode("start", Start{Name: "start"}, nil, 0)
	factory.AddNode("complete", Complete{Name: "complete"}, nil, 0)

	factory.AddNode("cutting1", CuttingMachine{Name: "cutting1"}, nil, 2*time.Second)
	factory.AddNode("cutting2", CuttingMachine{Name: "cutting2"}, nil, 2*time.Second)
	factory.AddNode("cutting3", CuttingMachine{Name: "cutting3"}, nil, 2*time.Second)

	factory.AddNode("sensor1", Sensor{Name: "sensor1"}, nil, time.Second)
	factory.AddNode("sensor2", Sensor{Name: "sensor2"}, nil, time.Second)
	factory.AddNode("sensor3", Sensor{Name: "sensor3"}, nil, time.Second)

	factory.AddNode("station1", Station{Name: "station1"}, stationMap(factory, "cutting1", "cutting2", "cutting3"), 0)
	factory.AddNode("station2", Station{Name: "station2"}, stationMap(factory, "sensor1", "sensor2", "sensor3"), 0)

	addEdges(factory, "start", "station1")
	addEdges(factory, "station1", "station2")
	addEdges(factory, "station2", "complete")
	addEdges(factory, "station2", "reject")

	return factory
}

func IntialiseConnections() map[string]*DataSources {
	connections := make(map[string]*DataSources)

	connections["cutting"] = &DataSources{Name: "cutting", DataType: Postgres, ConnectionDetails: os.Getenv("POSTGRES_URL"), ConnectionPoint: nil}
	connections["reject"] = &DataSources{Name: "reject", DataType: CSV, ConnectionDetails: "simData/test_data/reject.csv", ConnectionPoint: nil}
	connections["sensor"] = &DataSources{Name: "sensor", DataType: Kafka, ConnectionDetails: &KafkaInitializer{broker: os.Getenv("KAFKA_BROKER"), topic: "sensor"}, ConnectionPoint: nil}
	return connections
}

func CloseConnections(connections map[string]*DataSources) {
	for _, conn := range connections {
		switch conn.DataType {
		case CSV:
			if writer, ok := conn.ConnectionPoint.(*csv.Writer); ok {
				writer.Flush()
				if filepath, ok := conn.ConnectionDetails.(string); ok {
					os.Remove(filepath)
				}
			}
		case Postgres:
			if db, ok := conn.ConnectionPoint.(*sql.DB); ok {
				tableName := strings.ToLower(conn.Name)
				db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
				db.Close()
			}
		case Kafka:
			if producer, ok := conn.ConnectionPoint.(sarama.SyncProducer); ok {
				producer.Close()
			}
		}
	}
}

func SimulateData(connections map[string]*DataSources, factory *Factory) {
	defer CloseConnections(connections)
	var wg sync.WaitGroup

	createLogFile()
	totalNodes := len(factory.nodes)

	wg.Add(totalNodes + 1)
	for _, node := range factory.nodes {
		go node.Start(&wg, connections)
	}
	start := factory.GetNode("start")
	rate := 3
	go addParts(start, rate, &wg)
	wg.Wait()

	for _, node := range factory.nodes {
		close(node.Queue)
	}

	logging("All nodes have finished processing")
}

func createLogFile() {
	// timestamp := time.Now().Format("2006-01-02-15-04-05")
	// logPath = fmt.Sprintf("simData/log_data/log_%s.txt", timestamp)
	os.RemoveAll("simData/log_data")
	os.MkdirAll("simData/log_data", 0755)
	logPath = "simData/log_data/log.txt"
}
