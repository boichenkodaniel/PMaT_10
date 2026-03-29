package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type OrdersResponse struct {
	UserID int     `json:"user_id"`
	Orders []Order `json:"orders"`
}

type Order struct {
	OrderID int     `json:"order_id"`
	Product string  `json:"product"`
	Price   float64 `json:"price"`
	Status  string  `json:"status"`
}

type UserProfile struct {
	User   *User    `json:"user"`
	Orders []Order  `json:"orders"`
	Error  string   `json:"error,omitempty"`
}

var (
	userServiceURL  = &url.URL{Scheme: "http", Host: "localhost:8081"}
	orderServiceURL = &url.URL{Scheme: "http", Host: "localhost:8082"}
)

func reverseProxy(target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			targetQuery := target.RawQuery
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			req.Host = target.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			http.Error(w, `{"error": "service unavailable"}`, http.StatusServiceUnavailable)
		},
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	proxy := reverseProxy(userServiceURL)
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
	proxy.ServeHTTP(w, r)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	proxy := reverseProxy(orderServiceURL)
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
	proxy.ServeHTTP(w, r)
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
	}

	if userID == "" {
		http.Error(w, `{"error": "user id is required"}`, http.StatusBadRequest)
		return
	}

	var (
		user   *User
		orders []Order
		wg     sync.WaitGroup
		errMu  sync.Mutex
		errors []string
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := http.Get(fmt.Sprintf("http://localhost:8081/user?id=%s", userID))
		if err != nil {
			errMu.Lock()
			errors = append(errors, "user service unavailable")
			errMu.Unlock()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var u User
			if err := json.NewDecoder(resp.Body).Decode(&u); err == nil {
				user = &u
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := http.Get(fmt.Sprintf("http://localhost:8082/orders?user_id=%s", userID))
		if err != nil {
			errMu.Lock()
			errors = append(errors, "order service unavailable")
			errMu.Unlock()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var or OrdersResponse
			if err := json.NewDecoder(resp.Body).Decode(&or); err == nil {
				orders = or.Orders
			}
		}
	}()

	wg.Wait()

	profile := UserProfile{
		User:   user,
		Orders: orders,
	}

	if len(errors) > 0 {
		profile.Error = strings.Join(errors, "; ")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"status":        "healthy",
		"timestamp":     time.Now().Format(time.RFC3339),
		"user_service":  "checking...",
		"order_service": "checking...",
	}

	if resp, err := http.Get("http://localhost:8081/user?id=1"); err == nil {
		resp.Body.Close()
		status["user_service"] = "up"
	} else {
		status["user_service"] = "down"
	}

	if resp, err := http.Get("http://localhost:8082/orders?user_id=1"); err == nil {
		resp.Body.Close()
		status["order_service"] = "up"
	} else {
		status["order_service"] = "down"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/user", handleUser)
	mux.HandleFunc("/api/orders", handleOrders)
	mux.HandleFunc("/api/profile", handleProfile)
	mux.HandleFunc("/health", healthCheck)

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

	port := 8080
	fmt.Printf("API Gateway starting on port %d\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /api/user?id={id}     - Get user info")
	fmt.Println("  GET /api/orders?user_id={id} - Get user orders")
	fmt.Println("  GET /api/profile?id={id}  - Get user profile (info + orders)")
	fmt.Println("  GET /health               - Health check")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
