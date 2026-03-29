package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserServiceGetUser(t *testing.T) {
	req := httptest.NewRequest("GET", "/user?id=1", nil)
	w := httptest.NewRecorder()

	getUserHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var user User
	json.NewDecoder(resp.Body).Decode(&user)
	if user.ID != 1 {
		t.Errorf("Expected user ID 1, got %d", user.ID)
	}
	if user.Name == "" {
		t.Error("Expected user name")
	}
	if user.Phone == "" {
		t.Error("Expected user phone")
	}
}

func TestUserServiceGetUserMissingID(t *testing.T) {
	req := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	getUserHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserServiceGetUserInvalidID(t *testing.T) {
	req := httptest.NewRequest("GET", "/user?id=abc", nil)
	w := httptest.NewRecorder()

	getUserHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUserServiceGetUserNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/user?id=999", nil)
	w := httptest.NewRecorder()

	getUserHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestUserServiceGetUserAllUsers(t *testing.T) {
	testCases := []struct {
		id       string
		expected int
	}{
		{"1", 1},
		{"2", 2},
		{"3", 3},
	}

	for _, tc := range testCases {
		t.Run("user_"+tc.id, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/user?id="+tc.id, nil)
			w := httptest.NewRecorder()

			getUserHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			var user User
			json.NewDecoder(resp.Body).Decode(&user)
			if user.ID != tc.expected {
				t.Errorf("Expected user ID %d, got %d", tc.expected, user.ID)
			}
		})
	}
}
