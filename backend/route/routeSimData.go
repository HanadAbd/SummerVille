package route

import (
	"encoding/json"
	"fmt"
	"foo/backend/connections"
	"foo/simData"
	"net/http"
	"time"
)

type NodeGetRequest struct {
	Node_ID string `json:"node_id"`
}

type NodeSetRequest struct {
	NodeID         string   `json:"node_id"`
	NodeType       string   `json:"node_type"`
	NodesWithin    []string `json:"nodes_within"`
	NextNodes      []string `json:"next_nodes"`
	ProcessingTime float64  `json:"processing_time"`
	Event          string   `json:"event"`
}

type NodeSetResponse = simData.Node

type NodeGetResponse = simData.Node

func nodeGetRequestParams(r *http.Request) (*NodeGetRequest, error) {
	query := r.URL.Query()

	params := &NodeGetRequest{
		Node_ID: query.Get("node_id"),
	}

	return params, nil
}

func GetNode(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
	if r.Method != http.MethodGet {
		writeJSONErrorResponse(w, http.StatusMethodNotAllowed, "Only GET method is allowed")
		return
	}

	nodeGetRequestID, err := nodeGetRequestParams(r)
	if err != nil {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	factory := getFactory()
	if factory == nil {
		writeJSONErrorResponse(w, http.StatusInternalServerError, "Factory not found")
		return
	}

	data := factory.GetNodeData(nodeGetRequestID.Node_ID)
	if data == nil {
		writeJSONErrorResponse(w, http.StatusNotFound, "Node not found")
		return
	}
	message := fmt.Sprint("Node ", nodeGetRequestID.Node_ID, " found")
	writeJSONResponse(w, http.StatusOK, message, data)
}

func SetNode(w http.ResponseWriter, r *http.Request, prodConn *connections.ProdConn, connectors connections.WorkspaceConnectors) {
	if r.Method != http.MethodPost {
		writeJSONErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	var req NodeSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	factory := getFactory()
	if factory == nil {
		writeJSONErrorResponse(w, http.StatusInternalServerError, "Factory not found")
		return
	}
	nodesWithin := make(map[string]*simData.Node)
	for _, id := range req.NodesWithin {
		if node := factory.GetNode(id); node != nil {
			nodesWithin[id] = node
		}
	}

	nextNodes := make(map[string]*simData.Node)
	for _, id := range req.NextNodes {
		if node := factory.GetNode(id); node != nil {
			nextNodes[id] = node
		}
	}

	node := factory.GetNode(req.NodeID)
	if node == nil {
		writeJSONErrorResponse(w, http.StatusNotFound, "Node not found")
		return
	}

	node.NodesWithin = nodesWithin
	node.NextNodes = nextNodes
	node.ProcessingTime = time.Duration(req.ProcessingTime * float64(time.Second))

	message := fmt.Sprint("Node ", req.NodeID, " updated")
	writeJSONResponse(w, http.StatusOK, message, factory.GetNodeData(req.NodeID))
}
