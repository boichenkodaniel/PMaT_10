package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

var users = map[int]User{
	1: {ID: 1, Name: "Иван Петров", Phone: "+7-900-123-45-67"},
	2: {ID: 2, Name: "Анна Смирнова", Phone: "+7-900-987-65-43"},
	3: {ID: 3, Name: "Дмитрий Козлов", Phone: "+7-900-555-12-34"},
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error": "user id is required"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid user id"}`, http.StatusBadRequest)
		return
	}

	user, exists := users[id]
	if !exists {
		http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func main() {
	http.HandleFunc("/user", getUserHandler)

	port := 8081
	fmt.Printf("User service starting on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
