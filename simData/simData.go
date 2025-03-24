package simData

import (
	"context"
	"fmt"
	"foo/backend/connections"
	"foo/services/util"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var logPath = ""
var reg *util.Registry

func SetRegistry(registry *util.Registry) {
	reg = registry
}

type DataSource struct {
	Name       string
	DataType   string
	Table      *connections.TableDefinition
	Data       []interface{}
	Conditions func(*Node, *Part) bool
	DataMapper func(p *Part, n *Node) map[string]interface{}
}

type DataCondition struct {
	Field     string
	Operation string
	Value     interface{}
}

// EvaluateCondition checks if all conditions are met for a part

func (d *DataSource) Appender(p *Part, n *Node, dataPoints map[string]interface{}) {
	if d == nil {
		log.Printf("Warning: DataSource is nil")
		return
	}

	if d.Conditions == nil {
		log.Printf("Warning: Conditions function is nil for DataSource: %s", d.Name)
		return
	}

	if !d.Conditions(n, p) {
		return
	}

	if d.Table == nil {
		log.Printf("Warning: Table definition is nil for DataSource: %s", d.Name)
		return
	}

	if len(dataPoints) == 0 {
		log.Printf("No data points found for %s\n", d.Name)
		return
	}

	connectors := connections.GetWorkspaceConnectors()
	if connectors == nil {
		log.Printf("Error: No workspace connectors available for %s\n", d.Name)
		return
	}

	// Check if we have all required columns
	data := make([]interface{}, 0, 1) // Store a single map for the row
	rowData := make(map[string]interface{})
	missingColumns := []string{}

	for _, col := range d.Table.Columns {
		if val, exists := dataPoints[col.Name]; exists {
			rowData[col.Name] = val
		} else if !col.Nullable {
			missingColumns = append(missingColumns, col.Name)
		}
	}

	if len(missingColumns) > 0 {
		log.Printf("Error: Missing required columns for %s: %v\n", d.Name, missingColumns)
		return
	}

	// Only add the map once
	data = append(data, rowData)

	err := connectors.AddData(d.DataType, *d.Table, data)
	if err != nil {
		log.Printf("Failed to add data to %s: %v\n", d.Name, err)
	} else {
		log.Printf("Successfully added data to %s source\n", d.Name)
	}
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func SimulateData(connections map[string]*DataSource, factory *Factory, ctx context.Context) {
	defer CloseConnections()
	var wg sync.WaitGroup

	createLogFile()
	totalNodes := len(factory.nodes)
	simulationCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(totalNodes + 1)
	for _, node := range factory.nodes {
		go node.Start(&wg, connections, simulationCtx)
	}
	start := factory.GetNode("start")
	rate := 3
	go addParts(start, rate, &wg, simulationCtx)

	<-ctx.Done()
	log.Println("Simulation context cancelled, shutting down gracefully")

	cancel()

	for _, node := range factory.nodes {
		close(node.Queue)
	}

	wg.Wait()
	log.Println("All simulation goroutines have finished")
}

func getQueueContents(queue chan *Part) []string {
	contents := []string{}

	queueLen := len(queue)
	if queueLen == 0 {
		return contents
	}

	tempQueue := make([]string, 0, queueLen)

	for i := 0; i < queueLen; i++ {
		select {
		case part := <-queue:
			tempQueue = append(tempQueue, part.ID)
			queue <- part
		default:
			return tempQueue
		}
	}

	return tempQueue
}

func addParts(start *Node, rate int, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	counter := 0

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping part generation due to context cancellation")
			return
		default:
			counter++
			part := &Part{ID: "part" + fmt.Sprint(counter), Cutattempts: 0}

			select {
			case start.Queue <- part:
				log.Printf("Added new part: %s", part.ID)
			case <-time.After(500 * time.Millisecond):
				log.Printf("Queue full, skipping part: %s", part.ID)
			}

			duration := float64(time.Second) / ((rand.Float64() * float64(rate) / 2) + 0.75)
			time.Sleep(time.Duration(duration))
		}
	}
}

func CloseConnections() {
	connectors := connections.GetWorkspaceConnectors()
	if connectors != nil {
		connectors.Close()
	}
}

func createLogFile() {
	// timestamp := time.Now().Format("2006-01-02-15-04-05")
	// logPath = fmt.Sprintf("simData/log_data/log_%s.txt", timestamp)
	os.RemoveAll("simData/log_data")
	os.MkdirAll("simData/log_data", 0755)
	logPath = "simData/log_data/log.txt"
}
