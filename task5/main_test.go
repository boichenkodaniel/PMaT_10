package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestChatServerCreation(t *testing.T) {
	server := NewChatServer()

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if server.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}

	if server.register == nil {
		t.Error("Expected register channel to be initialized")
	}

	if server.unregister == nil {
		t.Error("Expected unregister channel to be initialized")
	}
}

func TestGetClientCount(t *testing.T) {
	server := NewChatServer()

	count := server.GetClientCount()
	if count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}

	client := &Client{
		nickname: "TestUser",
		send:     make(chan []byte, 256),
	}
	server.clients[client] = true

	count = server.GetClientCount()
	if count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}
}

func TestIsNicknameTaken(t *testing.T) {
	server := NewChatServer()

	if server.IsNicknameTaken("TestUser") {
		t.Error("Expected 'TestUser' to be available")
	}

	client := &Client{
		nickname: "TestUser",
		send:     make(chan []byte, 256),
	}
	server.clients[client] = true

	if !server.IsNicknameTaken("TestUser") {
		t.Error("Expected 'TestUser' to be taken")
	}

	if server.IsNicknameTaken("OtherUser") {
		t.Error("Expected 'OtherUser' to be available")
	}
}

func TestBroadcastToAll(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	time.Sleep(50 * time.Millisecond)

	client1 := &Client{
		nickname: "User1",
		send:     make(chan []byte, 256),
	}
	client2 := &Client{
		nickname: "User2",
		send:     make(chan []byte, 256),
	}

	server.register <- client1
	server.register <- client2

	time.Sleep(100 * time.Millisecond)

	for {
		select {
		case <-client1.send:
		case <-client2.send:
		default:
			goto doneDraining
		}
	}
doneDraining:

	testMsg := Message{
		Type:    MessageTypeChat,
		Content: "Hello everyone!",
	}
	server.BroadcastToAll(testMsg)

	received := 0
	timeout := time.After(2 * time.Second)

	for received < 2 {
		select {
		case msg := <-client1.send:
			var m Message
			json.Unmarshal(msg, &m)
			if m.Type == MessageTypeChat && m.Content != "Hello everyone!" {
				t.Errorf("Client1: Expected 'Hello everyone!', got '%s'", m.Content)
			}
			if m.Type == MessageTypeChat {
				received++
			}
		case msg := <-client2.send:
			var m Message
			json.Unmarshal(msg, &m)
			if m.Type == MessageTypeChat && m.Content != "Hello everyone!" {
				t.Errorf("Client2: Expected 'Hello everyone!', got '%s'", m.Content)
			}
			if m.Type == MessageTypeChat {
				received++
			}
		case <-timeout:
			t.Errorf("Timeout waiting for broadcast messages, received %d/2", received)
			return
		}
	}
}

func TestMessageMarshal(t *testing.T) {
	msg := Message{
		Type:        MessageTypeChat,
		Nickname:    "TestUser",
		Content:     "Hello!",
		Timestamp:   "2024-01-01T12:00:00Z",
		NewNickname: "",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.Type != msg.Type {
		t.Errorf("Expected type '%s', got '%s'", msg.Type, unmarshaled.Type)
	}
	if unmarshaled.Nickname != msg.Nickname {
		t.Errorf("Expected nickname '%s', got '%s'", msg.Nickname, unmarshaled.Nickname)
	}
	if unmarshaled.Content != msg.Content {
		t.Errorf("Expected content '%s', got '%s'", msg.Content, unmarshaled.Content)
	}
}

func TestWebSocketUpgrade(t *testing.T) {
	testServer := NewChatServer()
	go testServer.Run()

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleWebSocketWithServer(w, r, testServer)
	}))
	defer httpServer.Close()

	wsURL := "ws" + httpServer.URL[4:] + "/ws"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	var joinMsg Message
	err = ws.ReadJSON(&joinMsg)
	if err != nil {
		t.Fatalf("Failed to read join message: %v", err)
	}
	if joinMsg.Type != MessageTypeJoin {
		t.Errorf("Expected join message first, got '%s'", joinMsg.Type)
	}

	testMsg := Message{
		Type:    MessageTypeChat,
		Content: "Test message",
	}
	err = ws.WriteJSON(testMsg)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	var response Message
	err = ws.ReadJSON(&response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if response.Type != MessageTypeChat {
		t.Errorf("Expected message type '%s', got '%s'", MessageTypeChat, response.Type)
	}
	if response.Content != "Test message" {
		t.Errorf("Expected content 'Test message', got '%s'", response.Content)
	}
}

func TestClientCountEndpoint(t *testing.T) {
	testServer := NewChatServer()
	go testServer.Run()

	req := httptest.NewRequest("GET", "/clients", nil)
	w := httptest.NewRecorder()

	handleClientCountWithServer(w, req, testServer)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]int
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["clients"] != 0 {
		t.Errorf("Expected 0 clients, got %d", response["clients"])
	}
}

func TestGetCurrentTime(t *testing.T) {
	timeStr := getCurrentTime()

	_, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		t.Errorf("Time format is not RFC3339: %v", err)
	}
}

func TestRegisterClient(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		nickname: "NewUser",
		send:     make(chan []byte, 256),
	}

	server.register <- client

	time.Sleep(100 * time.Millisecond)

	count := server.GetClientCount()
	if count != 1 {
		t.Errorf("Expected 1 client after registration, got %d", count)
	}
}

func TestUnregisterClient(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	client := &Client{
		nickname: "TempUser",
		send:     make(chan []byte, 256),
	}

	server.register <- client
	time.Sleep(50 * time.Millisecond)

	server.unregister <- client
	time.Sleep(50 * time.Millisecond)

	count := server.GetClientCount()
	if count != 0 {
		t.Errorf("Expected 0 clients after unregistration, got %d", count)
	}
}

func TestJoinLeaveBroadcast(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	broadcastChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range server.broadcast {
			var m Message
			json.Unmarshal(msg, &m)
			broadcastChan <- m
		}
	}()

	client := &Client{
		nickname: "JoinTestUser",
		send:     make(chan []byte, 256),
	}
	server.register <- client

	select {
	case msg := <-broadcastChan:
		if msg.Type != MessageTypeJoin {
			t.Errorf("Expected join message, got '%s'", msg.Type)
		}
		if msg.Nickname != "JoinTestUser" {
			t.Errorf("Expected nickname 'JoinTestUser', got '%s'", msg.Nickname)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for join notification")
	}

	server.unregister <- client

	select {
	case msg := <-broadcastChan:
		if msg.Type != MessageTypeLeave {
			t.Errorf("Expected leave message, got '%s'", msg.Type)
		}
		if msg.Nickname != "JoinTestUser" {
			t.Errorf("Expected nickname 'JoinTestUser', got '%s'", msg.Nickname)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for leave notification")
	}

	close(server.broadcast)
	wg.Wait()
}

func TestNickChangeMessage(t *testing.T) {
	msg := Message{
		Type:        MessageTypeNickChange,
		Nickname:    "OldName",
		NewNickname: "NewName",
		Timestamp:   getCurrentTime(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled Message
	json.Unmarshal(data, &unmarshaled)

	if unmarshaled.Type != MessageTypeNickChange {
		t.Errorf("Expected type '%s', got '%s'", MessageTypeNickChange, unmarshaled.Type)
	}
	if unmarshaled.NewNickname != "NewName" {
		t.Errorf("Expected new nickname 'NewName', got '%s'", unmarshaled.NewNickname)
	}
}

func TestErrorMessage(t *testing.T) {
	msg := Message{
		Type:      MessageTypeError,
		Content:   "Test error message",
		Timestamp: getCurrentTime(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled Message
	json.Unmarshal(data, &unmarshaled)

	if unmarshaled.Type != MessageTypeError {
		t.Errorf("Expected type '%s', got '%s'", MessageTypeError, unmarshaled.Type)
	}
	if unmarshaled.Content != "Test error message" {
		t.Errorf("Expected content 'Test error message', got '%s'", unmarshaled.Content)
	}
}

func TestConcurrentClientAccess(t *testing.T) {
	server := NewChatServer()
	go server.Run()

	var wg sync.WaitGroup
	numClients := 10

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := &Client{
				nickname: "User",
				send:     make(chan []byte, 256),
			}
			server.register <- client
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	count := server.GetClientCount()
	if count != numClients {
		t.Errorf("Expected %d clients, got %d", numClients, count)
	}
}
