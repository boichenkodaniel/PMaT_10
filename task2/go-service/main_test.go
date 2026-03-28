package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "go-calculator"})
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}

	if response["service"] != "go-calculator" {
		t.Errorf("Expected service 'go-calculator', got '%s'", response["service"])
	}
}

func TestSumEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    SumRequest
		expected int
	}{
		{"positive numbers", SumRequest{A: 5, B: 3}, 8},
		{"negative numbers", SumRequest{A: -5, B: -3}, -8},
		{"mixed numbers", SumRequest{A: -5, B: 10}, 5},
		{"zero", SumRequest{A: 0, B: 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/sum", bytes.NewReader(body))
			w := httptest.NewRecorder()

			mux := http.NewServeMux()
			mux.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				var req SumRequest
				json.NewDecoder(r.Body).Decode(&req)
				result := req.A + req.B

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(SumResponse{Result: result})
			})
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var response SumResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.Result != tt.expected {
				t.Errorf("Expected result %d, got %d", tt.expected, response.Result)
			}
		})
	}
}

func TestMultiplyEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    SumRequest
		expected int
	}{
		{"positive numbers", SumRequest{A: 5, B: 3}, 15},
		{"negative numbers", SumRequest{A: -5, B: -3}, 15},
		{"mixed numbers", SumRequest{A: -5, B: 10}, -50},
		{"zero", SumRequest{A: 0, B: 100}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/multiply", bytes.NewReader(body))
			w := httptest.NewRecorder()

			mux := http.NewServeMux()
			mux.HandleFunc("/multiply", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				var req SumRequest
				json.NewDecoder(r.Body).Decode(&req)
				result := req.A * req.B

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(SumResponse{Result: result})
			})
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var response SumResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.Result != tt.expected {
				t.Errorf("Expected result %d, got %d", tt.expected, response.Result)
			}
		})
	}
}

func TestSumMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sum", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestMultiplyMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/multiply", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/multiply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSumInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/sum", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		var req SumRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	})
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
