package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGatewayHandleUser(t *testing.T) {
	origUserURL := userServiceURL
	origOrderURL := orderServiceURL
	defer func() {
		userServiceURL = origUserURL
		orderServiceURL = origOrderURL
	}()

	userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") == "1" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(User{ID: 1, Name: "Test User", Phone: "+7-000-000-00-00"})
		} else {
			http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
		}
	}))
	defer userServer.Close()

	orderServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer orderServer.Close()

	userServiceURL, _ = url.Parse(userServer.URL)
	orderServiceURL, _ = url.Parse(orderServer.URL)

	req := httptest.NewRequest("GET", "/user?id=1", nil)
	w := httptest.NewRecorder()

	handleUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var user User
	json.NewDecoder(resp.Body).Decode(&user)
	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
}

func TestGatewayHandleOrders(t *testing.T) {
	origUserURL := userServiceURL
	origOrderURL := orderServiceURL
	defer func() {
		userServiceURL = origUserURL
		orderServiceURL = origOrderURL
	}()

	userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer userServer.Close()

	orderServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("user_id") == "1" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(OrdersResponse{
				UserID: 1,
				Orders: []Order{{OrderID: 1, Product: "Test", Price: 100, Status: "new"}},
			})
		}
	}))
	defer orderServer.Close()

	userServiceURL, _ = url.Parse(userServer.URL)
	orderServiceURL, _ = url.Parse(orderServer.URL)

	req := httptest.NewRequest("GET", "/orders?user_id=1", nil)
	w := httptest.NewRecorder()

	handleOrders(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var ordersResp OrdersResponse
	json.NewDecoder(resp.Body).Decode(&ordersResp)
	if ordersResp.UserID != 1 {
		t.Errorf("Expected user_id 1, got %d", ordersResp.UserID)
	}
	if len(ordersResp.Orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(ordersResp.Orders))
	}
}

func TestGatewayHandleProfile(t *testing.T) {
	origUserURL := userServiceURL
	origOrderURL := orderServiceURL
	defer func() {
		userServiceURL = origUserURL
		orderServiceURL = origOrderURL
	}()

	userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(User{ID: 1, Name: "Test User", Phone: "+7-000-000-00-00"})
	}))
	defer userServer.Close()

	orderServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OrdersResponse{
			UserID: 1,
			Orders: []Order{{OrderID: 1, Product: "Test", Price: 100, Status: "new"}},
		})
	}))
	defer orderServer.Close()

	userServiceURL, _ = url.Parse(userServer.URL)
	orderServiceURL, _ = url.Parse(orderServer.URL)

	req := httptest.NewRequest("GET", "/profile?id=1", nil)
	w := httptest.NewRecorder()

	handleProfile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var profile UserProfile
	json.NewDecoder(resp.Body).Decode(&profile)
	if profile.User == nil {
		t.Error("Expected user data")
	}
	if profile.User.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", profile.User.ID)
	}
	if len(profile.Orders) == 0 {
		t.Error("Expected orders data")
	}
}

func TestGatewayHandleProfileMissingID(t *testing.T) {
	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	handleProfile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestGatewayHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthCheck(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var status map[string]string
	json.NewDecoder(resp.Body).Decode(&status)
	if status["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %s", status["status"])
	}
}

func TestGatewayRootHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]string{
			"name":        "API Gateway",
			"version":     "1.0.0",
			"endpoints":   "/api/user, /api/orders, /api/profile, /health",
			"description": "Routes requests to user and order microservices",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var info map[string]string
	json.NewDecoder(resp.Body).Decode(&info)
	if info["name"] != "API Gateway" {
		t.Errorf("Expected name 'API Gateway', got %s", info["name"])
	}
}
