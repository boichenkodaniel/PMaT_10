package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResponse struct {
	Result int `json:"result"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "go-calculator"})
	})

	mux.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SumRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		result := req.A + req.B

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SumResponse{Result: result})
	})

	mux.HandleFunc("/multiply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SumRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		result := req.A * req.B

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SumResponse{Result: result})
	})

	handler := LoggingMiddleware(mux)

	log.Println("Go Calculator Service starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
