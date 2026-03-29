package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MessageTypeChat       = "chat"
	MessageTypeJoin       = "join"
	MessageTypeLeave      = "leave"
	MessageTypeError      = "error"
	MessageTypeNickChange = "nick_change"
)

type Message struct {
	Type        string `json:"type"`
	Nickname    string `json:"nickname,omitempty"`
	Content     string `json:"content,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	NewNickname string `json:"new_nickname,omitempty"`
}

type Client struct {
	conn     *websocket.Conn
	nickname string
	send     chan []byte
}

type ChatServer struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *ChatServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()

			joinMsg := Message{
				Type:      MessageTypeJoin,
				Nickname:  client.nickname,
				Timestamp: getCurrentTime(),
			}
			joinData, _ := json.Marshal(joinMsg)
			s.broadcast <- joinData

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mu.Unlock()

			leaveMsg := Message{
				Type:      MessageTypeLeave,
				Nickname:  client.nickname,
				Timestamp: getCurrentTime(),
			}
			leaveData, _ := json.Marshal(leaveMsg)
			s.broadcast <- leaveData

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					log.Printf("Client %s buffer full, skipping message", client.nickname)
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *ChatServer) BroadcastToAll(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	s.broadcast <- data
}

func (s *ChatServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *ChatServer) IsNicknameTaken(nickname string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for client := range s.clients {
		if client.nickname == nickname {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var server *ChatServer

func getServer() *ChatServer {
	if server == nil {
		server = NewChatServer()
		go server.Run()
	}
	return server
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	handleWebSocketWithServer(w, r, getServer())
}

func handleWebSocketWithServer(w http.ResponseWriter, r *http.Request, srv *ChatServer) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:     conn,
		nickname: "Anonymous",
		send:     make(chan []byte, 256),
	}

	srv.register <- client

	go writePumpWithServer(client, srv)
	go readPumpWithServer(client, srv)
}

func writePump(client *Client) {
	writePumpWithServer(client, getServer())
}

func writePumpWithServer(client *Client, srv *ChatServer) {
	_ = srv
	defer func() {
		client.conn.Close()
	}()

	for message := range client.send {
		if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing to client %s: %v", client.nickname, err)
			return
		}
	}
}

func readPump(client *Client) {
	readPumpWithServer(client, getServer())
}

func readPumpWithServer(client *Client, srv *ChatServer) {
	defer func() {
		srv.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(512 * 1024)

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for client %s: %v", client.nickname, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			errorMsg := Message{
				Type:      MessageTypeError,
				Content:   "Invalid message format",
				Timestamp: getCurrentTime(),
			}
			srv.BroadcastToAll(errorMsg)
			continue
		}

		switch msg.Type {
		case MessageTypeChat:
			if msg.Content != "" {
				msg.Nickname = client.nickname
				msg.Timestamp = getCurrentTime()
				srv.BroadcastToAll(msg)
			}

		case MessageTypeNickChange:
			if msg.NewNickname != "" {
				if srv.IsNicknameTaken(msg.NewNickname) && msg.NewNickname != client.nickname {
					errorMsg := Message{
						Type:      MessageTypeError,
						Content:   fmt.Sprintf("Nickname '%s' is already taken", msg.NewNickname),
						Timestamp: getCurrentTime(),
					}
					errorData, _ := json.Marshal(errorMsg)
					select {
					case client.send <- errorData:
					default:
					}
				} else {
					oldNickname := client.nickname
					client.nickname = msg.NewNickname

					nickMsg := Message{
						Type:        MessageTypeNickChange,
						Nickname:    oldNickname,
						NewNickname: msg.NewNickname,
						Timestamp:   getCurrentTime(),
					}
					srv.BroadcastToAll(nickMsg)
				}
			}

		default:
			errorMsg := Message{
				Type:      MessageTypeError,
				Content:   fmt.Sprintf("Unknown message type: %s", msg.Type),
				Timestamp: getCurrentTime(),
			}
			srv.BroadcastToAll(errorMsg)
		}
	}
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}

func handleClientCount(w http.ResponseWriter, r *http.Request) {
	handleClientCountWithServer(w, r, getServer())
}

func handleClientCountWithServer(w http.ResponseWriter, r *http.Request, srv *ChatServer) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"clients": srv.GetClientCount(),
	})
}

func main() {
	server = NewChatServer()

	go server.Run()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/clients", handleClientCount)

	port := ":8080"
	fmt.Printf("Chat server starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
