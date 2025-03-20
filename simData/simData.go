package simData

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"foo/backend/connections"
	"foo/services/util"
	"log"
	"math/rand"
	"os"
	"path/filepath"
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

var logPath = ""
var reg *util.Registry

func SetRegistry(registry *util.Registry) {
	reg = registry
}

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
	Columns           []string
	Conditions        []DataCondition
	Frequency         int // How often data is recorded (in seconds, 0 = every time)
	LastUpdate        time.Time
	DataAdded         float64
	Registry          *util.Registry // Reference to the registry for service discovery
}

type DataCondition struct {
	Field     string
	Operation string
	Value     interface{}
}

// EvaluateCondition checks if all conditions are met for a part
func (d *DataSources) EvaluateCondition(p *Part) bool {
	if len(d.Conditions) == 0 {
		return true // No conditions means always record
	}

	for _, cond := range d.Conditions {
		switch cond.Field {
		case "CutVal":
			val := p.CutVal
			switch cond.Operation {
			case "eq":
				if val != cond.Value.(int) {
					return false
				}
			case "neq":
				if val == cond.Value.(int) {
					return false
				}
			case "gt":
				if val <= cond.Value.(int) {
					return false
				}
			case "lt":
				if val >= cond.Value.(int) {
					return false
				}
			}
		case "Cutattempts":
			val := p.Cutattempts
			switch cond.Operation {
			case "eq":
				if val != cond.Value.(int) {
					return false
				}
			case "neq":
				if val == cond.Value.(int) {
					return false
				}
			case "gt":
				if val <= cond.Value.(int) {
					return false
				}
			case "lt":
				if val >= cond.Value.(int) {
					return false
				}
			}
		}
	}
	return true
}

func (d *DataSources) ExtractData(p *Part, n *Node) []interface{} {
	row := make([]interface{}, 0, len(d.Columns))

	for _, col := range d.Columns {
		switch col {
		case "part_id":
			row = append(row, p.ID)
		case "cut_attempts":
			row = append(row, p.Cutattempts)
		case "cut_val":
			row = append(row, p.CutVal)
		case "time":
			row = append(row, time.Now())
		case "node_id":
			row = append(row, n.ID)
		case "machine_state":
			row = append(row, n.Event.String())
		}
	}

	return row
}

func (d *DataSources) Appender(p *Part, n *Node) {
	if !d.EvaluateCondition(p) {
		return
	}

	if d.Frequency > 0 {
		if time.Since(d.LastUpdate).Seconds() < float64(d.Frequency) {
			return
		}
	}

	if d.Data == nil {
		d.Data = &Data{
			Header: d.Columns,
			rows:   make([][]interface{}, 0),
		}

		// if _, err := d.InitializeDataSource(); err != nil {
		// 	logging("Error initializing data source %s: %v\n", d.Name, err)
		// 	return
		// }
		logging("Initialized %s\n", d.Name)
	}

	dataRow := d.ExtractData(p, n)

	data := &Data{
		Header: d.Columns,
		rows:   [][]interface{}{dataRow},
	}
	log.Println("Adding data to source")
	addToSource(d, data.rows)
	d.LastUpdate = time.Now()
	d.DataAdded++
}

func addToSource(c *DataSources, rows [][]interface{}) {
	tableDef := makeTableDefinition(c.Name, c.Columns)

	if c.ConnectionPoint == nil {
		log.Printf("No connection established for %s, trying to initialize", c.Name)
		return
	}

	switch c.DataType {
	case CSV:
		if conn, ok := c.ConnectionPoint.(*connections.CSVConn); ok {
			if err := conn.AddData(tableDef, convertRowsFormat(rows)); err != nil {
				logging("Error adding data to CSV %s: %v\n", c.Name, err)
			}
		} else {
			logging("Error: ConnectionPoint for %s is not a valid CSVConn\n", c.Name)
		}
	case Postgres:
		if conn, ok := c.ConnectionPoint.(*connections.PostgresConn); ok {
			if err := conn.AddData(tableDef, convertRowsFormat(rows)); err != nil {
				logging("Error adding data to Postgres %s: %v\n", c.Name, err)
			}
		} else {
			logging("Error: ConnectionPoint for %s is not a valid PostgresConn\n", c.Name)
		}
	case Kafka:
		if conn, ok := c.ConnectionPoint.(*connections.KafkaConn); ok {
			if err := conn.AddData(tableDef, convertRowsFormat(rows)); err != nil {
				logging("Error adding data to Kafka %s: %v\n", c.Name, err)
			}
		} else {
			logging("Error: ConnectionPoint for %s is not a valid KafkaConn\n", c.Name)
		}
	default:
		logging("Unsupported data type for %s\n", c.Name)
	}
}

func makeTableDefinition(name string, columns []string) connections.TableDefinition {
	schema := "public" // Default schema

	// Create column definitions from string names
	colDefs := make([]connections.ColumnDefinition, len(columns))
	for i, colName := range columns {
		colType := inferColumnType(colName)
		colDefs[i] = connections.ColumnDefinition{
			Name:     colName,
			Type:     connections.ColumnType(colType),
			Nullable: false,
		}
	}

	return connections.TableDefinition{
		Name:    name,
		Schema:  schema,
		Columns: colDefs,
	}
}

func inferColumnType(columnName string) string {
	columnName = strings.ToLower(columnName)

	if strings.Contains(columnName, "time") || strings.Contains(columnName, "date") {
		return "TIMESTAMP"
	} else if strings.Contains(columnName, "id") {
		return "TEXT"
	} else if strings.Contains(columnName, "val") ||
		strings.Contains(columnName, "count") ||
		strings.Contains(columnName, "attempts") ||
		strings.Contains(columnName, "number") {
		return "INT"
	} else if strings.Contains(columnName, "state") ||
		strings.Contains(columnName, "name") {
		return "TEXT"
	}

	return "TEXT"
}

func convertRowsFormat(rows [][]interface{}) []interface{} {
	result := make([]interface{}, len(rows))
	for i, row := range rows {
		result[i] = row
	}
	return result
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

			nextNode.Queue <- part

			n.Event = Idle
			logPartState(part.ID, "Idle", n.ID)

			queueContents := getQueueContents(n.Queue)
			if len(queueContents) > 0 {
				logNodeQueue(n.ID, queueContents)
			}
		}
	}
}

func IntialiseConnections(registry *util.Registry) map[string]*DataSources {
	conns := make(map[string]*DataSources)

	if err := os.MkdirAll("simData/test_data", 0755); err != nil {
		log.Printf("Error creating test_data directory: %v", err)
		return conns
	}

	metrics := &connections.ConnectionMetrics{
		OpenConnections: 5,
		IdleConnections: 2,
		QueryCount:      0,
		LastQueryTime:   0,
		Status:          "OK",
	}

	kafkaBroker := getEnvWithDefault("KAFKA_BROKER", "localhost:9092")

	postgresCred := &connections.PostgresCred{
		User:     getEnvWithDefault("PROD_DB_USER", "postgres"),
		Password: getEnvWithDefault("PROD_DB_PASSWORD", "postgres"),
		DBName:   getEnvWithDefault("PROD_DB_NAME", "postgres"),
		Host:     getEnvWithDefault("PROD_DB_HOST", "localhost"),
		Port:     getEnvWithDefault("PROD_DB_PORT", "5432"),
		SSLMode:  getEnvWithDefault("PROD_DB_SSLMODE", "disable"),
	}

	postgresConn, err := connections.InitPostgresDB(postgresCred, metrics)
	if err != nil {
		log.Printf("Error initializing Postgres connection: %v", err)
	}

	dataSourceConfigs := []struct {
		name       string
		dataType   DataType
		columns    []string
		conditions []DataCondition
		frequency  int
		connection interface{}
	}{
		{
			name:       "cutting",
			dataType:   Postgres,
			columns:    []string{"part_id", "cut_attempts", "cut_val", "time", "node_id"},
			conditions: []DataCondition{},
			frequency:  0,
			connection: &connections.PostgresConn{Conn: postgresConn, Name: "cutting"},
		},
		{
			name:     "reject",
			dataType: CSV,
			columns:  []string{"part_id", "cut_attempts", "cut_val", "time", "node_id"},
			conditions: []DataCondition{
				{Field: "CutVal", Operation: "lt", Value: 4},
				{Field: "CutVal", Operation: "gt", Value: 14},
			},
			frequency: 0,
			connection: &connections.CSVConn{
				Name:       "reject",
				FilePath:   filepath.Join("simData", "test_data", "reject.csv"),
				HasHeaders: true,
			},
		},
		{
			name:     "sensor",
			dataType: Kafka,
			columns:  []string{"part_id", "cut_attempts", "cut_val", "time", "machine_state"},
			conditions: []DataCondition{
				{Field: "Cutattempts", Operation: "gt", Value: 0},
			},
			frequency:  1,
			connection: nil,
		},
		// Add more data sources here easily
	}

	for _, config := range dataSourceConfigs {
		ds := &DataSources{
			Name:            config.name,
			DataType:        config.dataType,
			Data:            &Data{},
			Columns:         config.columns,
			Conditions:      config.conditions,
			Frequency:       config.frequency,
			Registry:        registry,
			ConnectionPoint: config.connection,
		}

		switch config.dataType {
		case CSV:
			if csvConn, ok := config.connection.(*connections.CSVConn); ok {
				ds.ConnectionDetails = &connections.CSVCredential{
					FilePath:  csvConn.FilePath,
					HasHeader: csvConn.HasHeaders,
					Encoding:  "utf-8",
				}
				if err := csvConn.InitCSV(); err != nil {
					log.Printf("Error initializing CSV connection %s: %v", config.name, err)
				}
			}
		case Postgres:
			ds.ConnectionDetails = postgresCred
		case Kafka:
			if kafkaConn, ok := config.connection.(*connections.KafkaConn); ok {
				ds.ConnectionDetails = &connections.KafkaCredential{
					Name:   config.name,
					Broker: kafkaBroker,
					Topic:  config.name,
				}
				if _, err := kafkaConn.InitKafka(ds.ConnectionDetails.(*connections.KafkaCredential),
					connections.ConnectionMetrics{Status: "initializing"}); err != nil {
					log.Printf("Error initializing Kafka connection %s: %v", config.name, err)
				}
			}
		}

		conns[config.name] = ds
	}

	if len(conns) == 0 {
		log.Println("Error: No connections initialized")
	} else {
		log.Printf("Successfully initialized %d data sources", len(conns))
	}
	log.Println("Data sources initialized")
	return conns
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
func stationMap(factory *Factory, names ...string) map[string]*Node {
	nodes := make(map[string]*Node)
	for _, name := range names {
		nodes[name] = factory.GetNode(name)
	}
	return nodes
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

func createLogFile() {
	// timestamp := time.Now().Format("2006-01-02-15-04-05")
	// logPath = fmt.Sprintf("simData/log_data/log_%s.txt", timestamp)
	os.RemoveAll("simData/log_data")
	os.MkdirAll("simData/log_data", 0755)
	logPath = "simData/log_data/log.txt"
}
