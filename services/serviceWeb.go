package services

import (
	"context"
	"foo/services/util"
	"foo/web"
	"log"
	"strings"
	"sync"

	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebService struct {
	mux       *http.ServeMux
	templates *template.Template
	server    *http.Server
	addr      string
	registry  *util.Registry

	upgrader     websocket.Upgrader
	hub          *Hub
	clientsMutex sync.Mutex
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	topics   map[string]bool
	registry *util.Registry
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func NewWebService(addr string, templates *template.Template, registry *util.Registry) *WebService {
	mux := http.NewServeMux()
	return &WebService{
		mux:       mux,
		templates: templates,
		addr:      addr,
		registry:  registry,

		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		hub: newHub(),
	}
}

func (s *WebService) Name() string {
	return "WebService"
}

func (s *WebService) Start(ctx context.Context) error {
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/dist"))))
	web.WebRouting(s.mux, s.templates, s.registry)

	go s.handleBroadcasts(ctx)
	s.mux.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server is running on http://%s/\n", s.addr)
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	}
}

func (s *WebService) GetMux() *http.ServeMux {
	return s.mux
}

func (s *WebService) handleBroadcasts(ctx context.Context) {
	for {
		select {
		case client := <-s.hub.register:
			s.hub.clients[client] = true
		case client := <-s.hub.unregister:
			if _, ok := s.hub.clients[client]; ok {
				delete(s.hub.clients, client)
				close(client.send)
			}

		case message := <-s.hub.broadcast:
			for client := range s.hub.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.hub.clients, client)
				}
			}

		case <-ctx.Done():
			return
		}
	}

}

func (s *WebService) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	topic := r.URL.Query().Get("topic")
	if topic == "" {
		topic = "default"
	}
	client := &Client{
		hub:      s.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		topics:   make(map[string]bool),
		registry: s.registry,
	}

	client.topics[topic] = true
	s.registry.RegisterWSChannel(topic, client.send)

	client.hub.register <- client

	go client.writePump()
	go client.readPump(s.registry)
}

func (c *Client) readPump(reg *util.Registry) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		// Unregister from all subscribed topics
		for topic := range c.topics {
			reg.UnregisterWSChannel(topic, c.send)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Handle subscription messages (simple implementation)
		// Example: "subscribe:logs" or "unsubscribe:logs"
		msgStr := string(message)
		if strings.HasPrefix(msgStr, "subscribe:") {
			topic := strings.TrimPrefix(msgStr, "subscribe:")
			c.topics[topic] = true
			reg.RegisterWSChannel(topic, c.send)
		} else if strings.HasPrefix(msgStr, "unsubscribe:") {
			topic := strings.TrimPrefix(msgStr, "unsubscribe:")
			delete(c.topics, topic)
			reg.UnregisterWSChannel(topic, c.send)
		}
	}
}

func (s *WebService) Stop(ctx context.Context) error {
	s.clientsMutex.Lock()
	for client, registered := range s.hub.clients {
		if registered {
			delete(s.hub.clients, client)
			close(client.send)
		}
	}
	s.clientsMutex.Unlock()
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
