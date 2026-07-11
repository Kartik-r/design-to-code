package main

import (
	"net/http"
	"os"

	"github.com/example/queue"
	"github.com/gin-gonic/gin"
)

func main() {
	brokerURL := os.Getenv("BROKER_URL")
	queue.Connect(brokerURL)

	router := gin.Default()
	router.POST("/orders", createOrder)
	router.Run(":8080")
}

func createOrder(c *gin.Context) {
	orderID := c.PostForm("order_id")
	queue.Publish("orders.created", orderID)
	c.JSON(http.StatusCreated, gin.H{"status": "queued"})
}