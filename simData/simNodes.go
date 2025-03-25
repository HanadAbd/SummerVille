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

type NodeVersion int

const (
	NodeTypeStart NodeVersion = iota
	NodeTypeReject
	NodeTypeComplete

	NodeTypeCuttingMachine
	NodeTypeWorker
	NodeTypeInventory
	NodeTypeSensorMachine
	NodeTypeRepairStation
	NodeTypeAssemblyStation
	NodeTypePackaging
	NodeTypeStation
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
	ID           string
	NodeHistory  []FactoryNode
	Cutattempts  int
	CutVal       int
	Weight       float64
	Temperature  float64
	Material     string
	DefectsCount int
	ProcessLog   []string
	IsPackaged   bool
	// Additional fields for demonstration
	TimesRepaired  int
	TimesAssembled int
	SensorReadings map[string]float64
}
type FactoryNode interface {
	GetID() string
	GetType() NodeVersion
	GetNodesWithin() map[string]FactoryNode
	GetNextNodes() map[string]FactoryNode
	GetQueue() chan *Part
	GetEvent() MachineState
	GetProcessingTime() time.Duration
	GetErrorNode() FactoryNode
	GetStation() FactoryNode

	SetID(string)
	SetType(NodeVersion)
	SetNodesWithin(map[string]FactoryNode)
	SetNextNodes(map[string]FactoryNode)
	SetQueue(chan *Part)
	SetEvent(MachineState)
	SetProcessingTime(time.Duration)
	SetErrorNode(FactoryNode)
	SetStation(FactoryNode)

	Type() NodeVersion
	Process(p *Part, c map[string]*DataSource) FactoryNode
	Start(wg *sync.WaitGroup, connections map[string]*DataSource, ctx context.Context)
	GetName() string
}

func (n *Node) GetName() string   { return "Node" }
func (n *Node) Type() NodeVersion { return NodeTypeStart }

func (n *Node) GetID() string                          { return n.ID }
func (n *Node) GetType() NodeVersion                   { return n.NodeVersion }
func (n *Node) GetNodesWithin() map[string]FactoryNode { return n.NodesWithin }
func (n *Node) GetNextNodes() map[string]FactoryNode   { return n.NextNodes }
func (n *Node) GetQueue() chan *Part                   { return n.Queue }
func (n *Node) GetEvent() MachineState                 { return n.Event }
func (n *Node) GetProcessingTime() time.Duration       { return n.ProcessingTime }
func (n *Node) GetErrorNode() FactoryNode              { return n.ErrorNode }
func (n *Node) GetStation() FactoryNode                { return n.Station }

func (n *Node) SetID(id string)                          { n.ID = id }
func (n *Node) SetType(t NodeVersion)                    { n.NodeVersion = t }
func (n *Node) SetNodesWithin(nw map[string]FactoryNode) { n.NodesWithin = nw }
func (n *Node) SetNextNodes(nn map[string]FactoryNode)   { n.NextNodes = nn }
func (n *Node) SetQueue(q chan *Part)                    { n.Queue = q }
func (n *Node) SetEvent(e MachineState)                  { n.Event = e }
func (n *Node) SetProcessingTime(pt time.Duration)       { n.ProcessingTime = pt }
func (n *Node) SetErrorNode(en FactoryNode)              { n.ErrorNode = en }
func (n *Node) SetStation(s FactoryNode)                 { n.Station = s }

func clearChannel(queue chan *Part) {
	for {
		select {
		case <-queue:

		default:
			return
		}
	}
}

type Node struct {
	ID             string
	NodeVersion    NodeVersion
	NodesWithin    map[string]FactoryNode
	NextNodes      map[string]FactoryNode
	Queue          chan *Part
	Event          MachineState
	ProcessingTime time.Duration
	ErrorNode      FactoryNode
	Station        FactoryNode

	Mu sync.Mutex
}

func (n *Node) Process(p *Part, c map[string]*DataSource) FactoryNode {
	if len(n.NextNodes) > 0 {
		for _, nextNode := range n.NextNodes {
			return nextNode
		}
	}
	return n.ErrorNode
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
			nextNode := processingPart(part, n, connections)

			if nextNode != nil {
				n.Event = Idle
				logPartState(part.ID, n.Event, n.ID)

				logPartTransition(part.ID, n.ID, nextNode.GetID())
				time.Sleep(time.Second)

				select {
				case nextNode.GetQueue() <- part:
					logPartState(part.ID, nextNode.GetEvent(), nextNode.GetID())

				case <-time.After(500 * time.Millisecond):
					log.Printf("Next node queue full, rejecting part %s", part.ID)
					n.ErrorNode.GetQueue() <- part
				case <-ctx.Done():
					log.Printf("Context cancelled while sending to next node, exiting %s", n.ID)
					return
				}

				queueLen := getQueueLength(n.Queue)
				if queueLen > 0 {
					logNodeQueue(n.ID, queueLen)
				}
			}

		default:
			if cancelled := noPartsAdded(ctx, n); cancelled {
				return
			}
		}
	}
}

func processingPart(part *Part, n FactoryNode, connections map[string]*DataSource) FactoryNode {
	n.SetEvent(Processing)

	logPartState(part.ID, n.GetEvent(), n.GetID())

	nextNode := n.Process(part, connections)

	time.Sleep(n.GetProcessingTime())
	part.NodeHistory = append(part.NodeHistory, n)

	n.SetEvent(Processed)
	logPartState(part.ID, n.GetEvent(), n.GetID())

	return nextNode
}

func noPartsAdded(ctx context.Context, n *Node) bool {
	if n.Event != Idle {
		n.Event = Idle
		logPartState("", n.Event, n.ID)
	}
	select {
	case <-time.After(100 * time.Millisecond):
		return false
	case <-ctx.Done():
		log.Printf("Context cancelled during idle, exiting node %s", n.ID)
		return true
	}
}

type Start struct {
	Node
	Name string
}

func (s *Start) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	s.Node.Mu.Lock()
	defer s.Node.Mu.Unlock()
	logPartState(p.ID, s.Event, s.ID)

	for _, node := range s.NextNodes {
		return node
	}
	return s.ErrorNode
}

func (s *Start) GetName() string {
	return "Start"
}

func (s *Start) Type() NodeVersion {
	return NodeTypeStart
}

type Reject struct {
	Node
	Name string
}

func (r *Reject) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	r.Node.Mu.Lock()
	defer r.Node.Mu.Unlock()

	logPartState(p.ID, r.Event, r.ID)

	if conn, exists := connections["reject"]; exists {
		conn.Appender(p, &r.Node, conn.DataMapper(p, &r.Node))
	}
	clearChannel(r.Queue)
	return nil
}

func (r *Reject) GetName() string {
	return "Reject"
}

func (r *Reject) Type() NodeVersion {
	return NodeTypeReject
}

type Complete struct {
	Node
	Name string
}

func (c *Complete) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	c.Node.Mu.Lock()
	defer c.Node.Mu.Unlock()
	logPartState(p.ID, c.Event, c.ID)

	if conn, exists := connections["complete"]; exists {
		conn.Appender(p, &c.Node, conn.DataMapper(p, &c.Node))
	}

	clearChannel(c.Queue)
	return nil
}

func (c *Complete) GetName() string {
	return "Complete"
}

func (c *Complete) Type() NodeVersion {
	return NodeTypeComplete
}

type Station struct {
	Node
	Name string
}

func (s *Station) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	logPartState(p.ID, s.Event, s.ID)

	childNodes := s.NodesWithin
	var rejectNode FactoryNode
	var exists bool
	if rejectNode, exists = s.GetNextNodes()["Reject"]; !exists || rejectNode == nil {
		rejectNode = s.ErrorNode
		if rejectNode == nil && len(s.NextNodes) > 0 {
			for _, node := range s.NextNodes {
				rejectNode = node
				break
			}
		}
		if rejectNode == nil {
			logging("Warning: No reject node found for station %s and no fallback available\n", s.ID)
			return nil
		}
	}

	earliestNode := rejectNode

	for _, node := range childNodes {
		if len(node.GetNextNodes()) == 0 {
			node.SetNextNodes(s.NextNodes)
		}

		if node.GetEvent() == Idle {
			return node
		}

		if earliestNode == nil || len(node.GetQueue()) < len(earliestNode.GetQueue()) {
			earliestNode = node
		}
	}

	if earliestNode != nil && earliestNode != rejectNode {
		return earliestNode
	}

	return s.ErrorNode
}

func (c *Station) GetName() string {
	return "Station"
}

func (c *Station) Type() NodeVersion {
	return NodeTypeStation
}

type CuttingMachineNode struct {
	Node
	Name                string
	FailureRate         float64
	TimeSinceLastRepair time.Duration
	Dullness            float64
	Tools               []string
}

func NewCuttingMachineNode(id string, processingTime time.Duration) *CuttingMachineNode {
	return &CuttingMachineNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypeCuttingMachine,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		FailureRate: 0.01,
		Dullness:    0.0,
		Tools:       []string{"SteelBlade"},
	}
}

func (cm *CuttingMachineNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	cm.Node.Mu.Lock()
	defer cm.Node.Mu.Unlock()

	logPartState(p.ID, cm.Event, cm.ID)

	cm.TimeSinceLastRepair += cm.ProcessingTime

	if cm.ErrorNode != nil && rand.Float64() < cm.FailureRate {
		return cm.ErrorNode
	}

	if rand.Float64() < (0.1 + cm.Dullness*0.05) {
		p.DefectsCount++
	}
	for _, node := range cm.NextNodes {
		return node
	}
	return cm.ErrorNode
}

func (cm *CuttingMachineNode) GetName() string {
	return "Cutting Machine"
}

func (cm *CuttingMachineNode) Type() NodeVersion {
	return NodeTypeCuttingMachine
}

type WorkerNode struct {
	Node
	Name       string
	Department string
	SkillLevel int
}

func NewWorkerNode(id, name, dept string, skill int, processingTime time.Duration) *WorkerNode {
	return &WorkerNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypeWorker,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		Name:       name,
		Department: dept,
		SkillLevel: skill,
	}
}

func (w *WorkerNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	logPartState(p.ID, w.Event, w.ID)

	if p.DefectsCount > 0 {
		fixChance := float64(w.SkillLevel) * 0.05
		if rand.Float64() < fixChance {
			p.DefectsCount--
			logging("Worker %s fixed a defect on part %s\n", w.ID, p.ID)
		} else {
			logging("Worker %s could not fix a defect on part %s\n", w.ID, p.ID)
			return w.GetNextNodes()["Reject"]
		}
	} else {
		logging("No defects found on part %s\n", p.ID)
	}
	for _, node := range w.NextNodes {
		return node
	}
	return w.ErrorNode
}

func (w *WorkerNode) GetName() string {
	return "Worker"
}

func (w *WorkerNode) Type() NodeVersion {
	return NodeTypeWorker
}

type Sensor struct {
	Name    string
	Station *Node
}

func (s Sensor) Process(n *Node, p *Part, connections map[string]*DataSource) FactoryNode {
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

func (s Sensor) Type() NodeVersion {
	return NodeTypeSensorMachine
}

type InventoryNode struct {
	Node
	Name          string
	Capacity      int
	StoredParts   []*Part
	AllowedTypes  []string
	CurrentStored int
}

func NewInventoryNode(id string, capacity int, allowedTypes []string, processingTime time.Duration) *InventoryNode {
	return &InventoryNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypeInventory,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		Capacity:     capacity,
		AllowedTypes: allowedTypes,
		StoredParts:  []*Part{},
	}
}

func (inv *InventoryNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	inv.Mu.Lock()
	defer inv.Mu.Unlock()

	logPartState(p.ID, inv.Event, inv.ID)

	// Check if this part type is allowed
	allowed := false
	for _, t := range inv.AllowedTypes {
		if p.Material == t { // Using Material as PartType equivalent
			allowed = true
			break
		}
	}

	if !allowed {
		logging("Part type %s not allowed in inventory %s\n", p.Material, inv.ID)
		return inv.ErrorNode
	}

	// Check capacity
	if inv.CurrentStored >= inv.Capacity {
		logging("Inventory %s is full! Cannot store part %s\n", inv.ID, p.ID)
		return inv.ErrorNode
	}

	// Store the part
	inv.StoredParts = append(inv.StoredParts, p)
	inv.CurrentStored++
	logging("Part %s stored in inventory %s\n", p.ID, inv.ID)

	// Pass to next node
	for _, node := range inv.NextNodes {
		return node
	}
	return inv.ErrorNode
}

func (inv *InventoryNode) GetName() string {
	return "Inventory"
}

func (inv *InventoryNode) Type() NodeVersion {
	return NodeTypeInventory
}

type SensorMachineNode struct {
	Node
	Name          string
	Calibration   float64
	FailureChance float64
}

func NewSensorMachineNode(id string, processingTime time.Duration) *SensorMachineNode {
	return &SensorMachineNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypeSensorMachine,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		Calibration:   1.0,
		FailureChance: 0.02,
	}
}

func (s *SensorMachineNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	logPartState(p.ID, s.Event, s.ID)

	// Check for failure
	if rand.Float64() < s.FailureChance {
		logging("Sensor machine %s FAILED mid-scan for part %s!\n", s.ID, p.ID)
		return s.ErrorNode
	}

	// Perform sensor reading (fake)
	if p.SensorReadings == nil {
		p.SensorReadings = make(map[string]float64)
	}
	dimension := 100.0 + rand.Float64()*5.0
	p.SensorReadings["dimension"] = dimension * s.Calibration
	logging("Sensor %s reading: dimension=%.2f for part %s\n", s.ID, p.SensorReadings["dimension"], p.ID)

	// Add data to the sensor data source if available
	if conn, exists := connections["sensor_readings"]; exists {
		conn.Appender(p, &s.Node, conn.DataMapper(p, &s.Node))
	}

	// Pass to next node
	for _, node := range s.NextNodes {
		return node
	}
	return s.ErrorNode
}

func (s *SensorMachineNode) GetName() string {
	return "Sensor Machine"
}

func (s *SensorMachineNode) Type() NodeVersion {
	return NodeTypeSensorMachine
}

type RepairStationNode struct {
	Node
	Name           string
	RepairCapacity int
}

func NewRepairStationNode(id string, processingTime time.Duration, capacity int) *RepairStationNode {
	return &RepairStationNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypeRepairStation,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		RepairCapacity: capacity,
	}
}

func (rs *RepairStationNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()

	logPartState(p.ID, rs.Event, rs.ID)
	logging("Repair station %s checking for defects on part %s\n", rs.ID, p.ID)

	if p.DefectsCount > 0 {
		// Fix as many defects as possible up to capacity
		defectsFixed := p.DefectsCount
		if defectsFixed > rs.RepairCapacity {
			defectsFixed = rs.RepairCapacity
		}

		p.DefectsCount -= defectsFixed
		p.TimesRepaired++

		if p.DefectsCount == 0 {
			logging("Repair station %s fixed all %d defects on part %s\n", rs.ID, defectsFixed, p.ID)
		} else {
			logging("Repair station %s fixed %d defects, %d remain on part %s\n",
				rs.ID, defectsFixed, p.DefectsCount, p.ID)
		}
	} else {
		logging("Repair station %s found no defects on part %s\n", rs.ID, p.ID)
	}

	// Pass to next node
	for _, node := range rs.NextNodes {
		return node
	}
	return rs.ErrorNode
}

func (rs *RepairStationNode) GetName() string {
	return "Repair Station"
}

func (rs *RepairStationNode) Type() NodeVersion {
	return NodeTypeRepairStation
}

type AssemblyStationNode struct {
	Node
	Name          string
	ToolsRequired []string
}

func NewAssemblyStationNode(id string, processingTime time.Duration, tools []string) *AssemblyStationNode {
	return &AssemblyStationNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypeAssemblyStation,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		ToolsRequired: tools,
	}
}

func (as *AssemblyStationNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	as.Mu.Lock()
	defer as.Mu.Unlock()

	logPartState(p.ID, as.Event, as.ID)
	logging("Assembly station %s combining components for part %s\n", as.ID, p.ID)

	// Example logic: each assembly step might add weight
	p.Weight += 2.0
	p.TimesAssembled++

	// Could implement tool check logic here
	hasAllTools := true
	if hasAllTools {
		logging("Assembly successful using tools %v for part %s\n", as.ToolsRequired, p.ID)
	} else {
		logging("Assembly incomplete due to missing tools for part %s\n", p.ID)
		return as.ErrorNode
	}

	// Pass to next node
	for _, node := range as.NextNodes {
		return node
	}
	return as.ErrorNode
}

func (as *AssemblyStationNode) GetName() string {
	return "Assembly Station"
}

func (as *AssemblyStationNode) Type() NodeVersion {
	return NodeTypeAssemblyStation
}

type PackagingNode struct {
	Node
	Name          string
	PackagingType string
}

func NewPackagingNode(id, pType string, processingTime time.Duration) *PackagingNode {
	return &PackagingNode{
		Node: Node{
			ID:             id,
			NodeVersion:    NodeTypePackaging,
			Queue:          make(chan *Part),
			ProcessingTime: processingTime,
		},
		PackagingType: pType,
	}
}

func (pack *PackagingNode) Process(p *Part, connections map[string]*DataSource) FactoryNode {
	pack.Mu.Lock()
	defer pack.Mu.Unlock()

	logPartState(p.ID, pack.Event, pack.ID)

	p.IsPackaged = true
	logging("Part %s packaged using %s\n", p.ID, pack.PackagingType)

	// Add packaging data to data source if available
	if conn, exists := connections["packaging"]; exists {
		conn.Appender(p, &pack.Node, conn.DataMapper(p, &pack.Node))
	}

	// Pass to next node (usually Complete node)
	for _, node := range pack.NextNodes {
		return node
	}
	return pack.ErrorNode
}

func (pack *PackagingNode) GetName() string {
	return "Packaging"
}

func (pack *PackagingNode) Type() NodeVersion {
	return NodeTypePackaging
}
func IntialiseConnections(registry *util.Registry) map[string]*DataSource {
	conns := make(map[string]*DataSource)

	// Machine operations data source
	conns["machine_operations"] = &DataSource{
		Name:     "machine_operations",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "machine_operations",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "machine_id", Type: connections.TypeText, Nullable: false},
				{Name: "operation_type", Type: connections.TypeText, Nullable: false},
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "duration_seconds", Type: connections.TypeFloat, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			// All machine nodes except storage
			nodeType := n.NodeVersion
			return nodeType == NodeTypeCuttingMachine ||
				nodeType == NodeTypeSensorMachine ||
				nodeType == NodeTypeRepairStation ||
				nodeType == NodeTypeAssemblyStation ||
				nodeType == NodeTypePackaging
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			var operationType string
			switch n.GetType() {
			case NodeTypeCuttingMachine:
				operationType = "Cutting"
			case NodeTypeSensorMachine:
				operationType = "Quality Check"
			case NodeTypeRepairStation:
				operationType = "Repair"
			case NodeTypeAssemblyStation:
				operationType = "Assembly"
			case NodeTypePackaging:
				operationType = "Packaging"
			default:
				operationType = "Unknown"
			}

			return map[string]interface{}{
				"machine_id":       n.GetID(),
				"operation_type":   operationType,
				"part_id":          p.ID,
				"duration_seconds": n.GetProcessingTime().Seconds(),
				"timestamp":        time.Now(),
			}
		},
	}

	// Worker activity data source
	conns["worker_activity"] = &DataSource{
		Name:     "worker_activity",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "worker_activity",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "worker_id", Type: connections.TypeText, Nullable: false},
				{Name: "department", Type: connections.TypeText, Nullable: false},
				{Name: "activity", Type: connections.TypeText, Nullable: false},
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "skill_level", Type: connections.TypeInt, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return n.NodeVersion == NodeTypeWorker
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			worker, ok := n.(*WorkerNode)
			activity := "Processing"
			skillLevel := 1
			department := "Unknown"

			if ok {
				department = worker.Department
				skillLevel = worker.SkillLevel

				if p.DefectsCount > 0 {
					activity = "Repairing"
				}
			}

			return map[string]interface{}{
				"worker_id":   n.GetID(),
				"department":  department,
				"activity":    activity,
				"part_id":     p.ID,
				"skill_level": skillLevel,
				"timestamp":   time.Now(),
			}
		},
	}

	// Inventory tracking
	conns["inventory_tracking"] = &DataSource{
		Name:     "inventory_tracking",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "inventory_tracking",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "inventory_id", Type: connections.TypeText, Nullable: false},
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "action", Type: connections.TypeText, Nullable: false},
				{Name: "current_stored", Type: connections.TypeInt, Nullable: false},
				{Name: "max_capacity", Type: connections.TypeInt, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return n.NodeVersion == NodeTypeInventory
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			inventory, ok := n.(*InventoryNode)
			currentStored := 0
			maxCapacity := 0

			if ok {
				currentStored = inventory.CurrentStored
				maxCapacity = inventory.Capacity
			}

			return map[string]interface{}{
				"inventory_id":   n.GetID(),
				"part_id":        p.ID,
				"action":         "Store",
				"current_stored": currentStored,
				"max_capacity":   maxCapacity,
				"timestamp":      time.Now(),
			}
		},
	}

	// Quality control data
	conns["quality_control"] = &DataSource{
		Name:     "quality_control",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "quality_measurements",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "sensor_id", Type: connections.TypeText, Nullable: false},
				{Name: "measurement_type", Type: connections.TypeText, Nullable: false},
				{Name: "measurement_value", Type: connections.TypeFloat, Nullable: false},
				{Name: "within_spec", Type: connections.TypeBoolean, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return n.NodeVersion == NodeTypeSensorMachine && p.SensorReadings != nil
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			sensor, ok := n.(*SensorMachineNode)
			measurementType := "Unknown"
			measurementValue := 0.0
			withinSpec := true

			if ok && len(p.SensorReadings) > 0 {
				if strings.Contains(sensor.Name, "Dimension") {
					measurementType = "Dimension"
				} else if strings.Contains(sensor.Name, "Surface") {
					measurementType = "Surface"
				} else if strings.Contains(sensor.Name, "Weight") {
					measurementType = "Weight"
				} else if strings.Contains(sensor.Name, "Inspection") {
					measurementType = "Final Inspection"
				}

				for _, value := range p.SensorReadings {
					measurementValue = value
					break
				}

				if measurementType == "Dimension" {
					withinSpec = measurementValue >= 98.0 && measurementValue <= 105.0
				} else if measurementType == "Weight" {
					withinSpec = measurementValue >= 95.0 && measurementValue <= 105.0
				}
			}

			return map[string]interface{}{
				"part_id":           p.ID,
				"sensor_id":         n.GetID(),
				"measurement_type":  measurementType,
				"measurement_value": measurementValue,
				"within_spec":       withinSpec,
				"timestamp":         time.Now(),
			}
		},
	}

	conns["defect_tracking"] = &DataSource{
		Name:     "defect_tracking",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "defects",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "defect_count", Type: connections.TypeInt, Nullable: false},
				{Name: "detected_at", Type: connections.TypeText, Nullable: false},
				{Name: "times_repaired", Type: connections.TypeInt, Nullable: false},
				{Name: "repairable", Type: connections.TypeBoolean, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return p.DefectsCount > 0
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			return map[string]interface{}{
				"part_id":        p.ID,
				"defect_count":   p.DefectsCount,
				"detected_at":    n.GetID(),
				"times_repaired": p.TimesRepaired,
				"repairable":     p.DefectsCount <= 3,
				"timestamp":      time.Now(),
			}
		},
	}

	conns["part_completion"] = &DataSource{
		Name:     "part_completion",
		DataType: "postgres",
		Table: &connections.TableDefinition{
			Name:   "part_status",
			Schema: "test",
			Columns: []connections.ColumnDefinition{
				{Name: "part_id", Type: connections.TypeText, Nullable: false},
				{Name: "status", Type: connections.TypeText, Nullable: false},
				{Name: "total_processing_time", Type: connections.TypeFloat, Nullable: false},
				{Name: "is_packaged", Type: connections.TypeBoolean, Nullable: false},
				{Name: "station_id", Type: connections.TypeText, Nullable: false},
				{Name: "timestamp", Type: connections.TypeDate, Nullable: false},
			},
		},
		Conditions: func(n *Node, p *Part) bool {
			return n.ID == "complete" || n.ID == "reject"
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			status := "Completed"
			if n.GetID() == "reject" {
				status = "Rejected"
			}

			return map[string]interface{}{
				"part_id":               p.ID,
				"status":                status,
				"total_processing_time": float64(len(p.NodeHistory)) * 1.5,
				"is_packaged":           p.IsPackaged,
				"station_id":            n.GetID(),
				"timestamp":             time.Now(),
			}
		},
	}

	// Also keep the original data sources
	conns["cutting"] = &DataSource{
		Name:     "cutting",
		DataType: "postgres",
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
			return n.NodeVersion == NodeTypeCuttingMachine
		},
		DataMapper: func(p *Part, n FactoryNode) map[string]interface{} {
			return map[string]interface{}{
				"part_id":      p.ID,
				"cut_attempts": p.Cutattempts,
				"cut_val":      p.CutVal,
				"timestamp":    time.Now(),
			}
		},
	}

	log.Println("Data sources initialized")
	return conns
}
