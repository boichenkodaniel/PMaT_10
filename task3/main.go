package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

type Item struct {
	ID          *int     `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Price       float64  `json:"price"`
}

type ItemResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
}

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
	Length  int    `json:"length"`
}

var (
	itemsDB  = make(map[int]Item)
	nextID   = 1
	itemsMux sync.RWMutex
)

func main() {
	r := gin.Default()

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func parseParam(c *gin.Context, key string) (int, error) {
	var id int
	_, err := fmt.Sscanf(c.Param(key), "%d", &id)
	return id, err
}
