package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	itemsDB = make(map[int]Item)
	nextID = 1

	t.Cleanup(func() {
		itemsDB = make(map[int]Item)
		nextID = 1
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Task3 API",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	r.POST("/echo", func(c *gin.Context) {
		var req EchoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		c.JSON(http.StatusOK, EchoResponse{
			Message: req.Message,
			Length:  len(req.Message),
		})
	})

	r.GET("/items", func(c *gin.Context) {
		itemsMux.RLock()
		defer itemsMux.RUnlock()

		result := make([]ItemResponse, 0, len(itemsDB))
		for _, item := range itemsDB {
			result = append(result, ItemResponse{
				ID:          *item.ID,
				Name:        item.Name,
				Description: item.Description,
				Price:       item.Price,
			})
		}
		c.JSON(http.StatusOK, result)
	})

	r.POST("/items", func(c *gin.Context) {
		var item Item
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		itemsMux.Lock()
		newID := nextID
		nextID++
		newItem := Item{
			ID:          &newID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
		}
		itemsDB[newID] = newItem
		itemsMux.Unlock()

		c.JSON(http.StatusCreated, ItemResponse{
			ID:          *newItem.ID,
			Name:        newItem.Name,
			Description: newItem.Description,
			Price:       newItem.Price,
		})
	})

	r.GET("/items/:id", func(c *gin.Context) {
		id, err := parseParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}

		itemsMux.RLock()
		item, exists := itemsDB[id]
		itemsMux.RUnlock()

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}

		c.JSON(http.StatusOK, ItemResponse{
			ID:          *item.ID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
		})
	})

	r.PUT("/items/:id", func(c *gin.Context) {
		id, err := parseParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}

		var item Item
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		itemsMux.Lock()
		_, exists := itemsDB[id]
		if !exists {
			itemsMux.Unlock()
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}

		updatedItem := Item{
			ID:          &id,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
		}
		itemsDB[id] = updatedItem
		itemsMux.Unlock()

		c.JSON(http.StatusOK, ItemResponse{
			ID:          *updatedItem.ID,
			Name:        updatedItem.Name,
			Description: updatedItem.Description,
			Price:       updatedItem.Price,
		})
	})

	r.DELETE("/items/:id", func(c *gin.Context) {
		id, err := parseParam(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}

		itemsMux.Lock()
		_, exists := itemsDB[id]
		if !exists {
			itemsMux.Unlock()
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		delete(itemsDB, id)
		itemsMux.Unlock()

		c.Status(http.StatusNoContent)
	})

	return r
}

func TestRoot(t *testing.T) {
	r := setupTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Welcome to Task3 API", response["message"])
}

func TestHealthCheck(t *testing.T) {
	r := setupTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}

func TestEcho(t *testing.T) {
	r := setupTestRouter(t)

	t.Run("returns message with length", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"message": "Hello World"}`)
		req, _ := http.NewRequest("POST", "/echo", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response EchoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Hello World", response.Message)
		assert.Equal(t, 11, response.Length)
	})

	t.Run("empty message", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"message": ""}`)
		req, _ := http.NewRequest("POST", "/echo", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response EchoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "", response.Message)
		assert.Equal(t, 0, response.Length)
	})

	t.Run("long message", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"message": "` + string(bytes.Repeat([]byte("A"), 1000)) + `"}`)
		req, _ := http.NewRequest("POST", "/echo", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response EchoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 1000, response.Length)
	})
}

func TestItems(t *testing.T) {
	r := setupTestRouter(t)

	t.Run("get items empty", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 0)
	})

	t.Run("create item", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"name": "Test Item", "description": "Test Description", "price": 19.99}`)
		req, _ := http.NewRequest("POST", "/items", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 1, response.ID)
		assert.Equal(t, "Test Item", response.Name)
		assert.Equal(t, "Test Description", *response.Description)
		assert.Equal(t, 19.99, response.Price)
	})

	t.Run("create item without description", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"name": "Test Item", "price": 29.99}`)
		req, _ := http.NewRequest("POST", "/items", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 1, response.ID)
		assert.Equal(t, "Test Item", response.Name)
		assert.Nil(t, response.Description)
		assert.Equal(t, 29.99, response.Price)
	})

	t.Run("get item by id", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		createW := httptest.NewRecorder()
		createBody := bytes.NewBufferString(`{"name": "Item 1", "price": 10.0}`)
		createReq, _ := http.NewRequest("POST", "/items", createBody)
		createReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(createW, createReq)

		var createdItem ItemResponse
		json.Unmarshal(createW.Body.Bytes(), &createdItem)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, createdItem.ID, response.ID)
		assert.Equal(t, "Item 1", response.Name)
		assert.Equal(t, 10.0, response.Price)
	})

	t.Run("get item not found", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items/9999", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update item", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		createW := httptest.NewRecorder()
		createBody := bytes.NewBufferString(`{"name": "Original", "price": 10.0}`)
		createReq, _ := http.NewRequest("POST", "/items", createBody)
		createReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(createW, createReq)

		updateW := httptest.NewRecorder()
		updateBody := bytes.NewBufferString(`{"name": "Updated", "price": 20.0}`)
		updateReq, _ := http.NewRequest("PUT", "/items/1", updateBody)
		updateReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(updateW, updateReq)

		assert.Equal(t, http.StatusOK, updateW.Code)
		var response ItemResponse
		json.Unmarshal(updateW.Body.Bytes(), &response)
		assert.Equal(t, 1, response.ID)
		assert.Equal(t, "Updated", response.Name)
		assert.Equal(t, 20.0, response.Price)
	})

	t.Run("update item not found", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"name": "Test", "price": 10.0}`)
		req, _ := http.NewRequest("PUT", "/items/9999", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete item", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		createW := httptest.NewRecorder()
		createBody := bytes.NewBufferString(`{"name": "To Delete", "price": 5.0}`)
		createReq, _ := http.NewRequest("POST", "/items", createBody)
		createReq.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(createW, createReq)

		deleteW := httptest.NewRecorder()
		deleteReq, _ := http.NewRequest("DELETE", "/items/1", nil)
		r.ServeHTTP(deleteW, deleteReq)

		assert.Equal(t, http.StatusNoContent, deleteW.Code)

		getW := httptest.NewRecorder()
		getReq, _ := http.NewRequest("GET", "/items/1", nil)
		r.ServeHTTP(getW, getReq)
		assert.Equal(t, http.StatusNotFound, getW.Code)
	})

	t.Run("delete item not found", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/items/9999", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get multiple items", func(t *testing.T) {
		itemsDB = make(map[int]Item)
		nextID = 1
		for i := 0; i < 3; i++ {
			createW := httptest.NewRecorder()
			createBody := bytes.NewBufferString(`{"name": "Item ` + string(rune('0'+i)) + `", "price": ` + string(rune('0'+i)) + `.0}`)
			createReq, _ := http.NewRequest("POST", "/items", createBody)
			createReq.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(createW, createReq)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 3)
	})
}

func TestItemValidation(t *testing.T) {
	r := setupTestRouter(t)

	t.Run("create item with zero price", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"name": "Free Item", "price": 0.0}`)
		req, _ := http.NewRequest("POST", "/items", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 0.0, response.Price)
	})

	t.Run("create item with negative price", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"name": "Negative Item", "price": -10.0}`)
		req, _ := http.NewRequest("POST", "/items", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response ItemResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, -10.0, response.Price)
	})
}
