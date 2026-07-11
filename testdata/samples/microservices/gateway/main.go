package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/aggregate", aggregateHandler)
	router.GET("/health", healthCheck)

	router.Run(":8080")
}

func aggregateHandler(c *gin.Context) {
	userResp, err := http.Get("http://service-a:9001/users")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "service-a unavailable"})
		return
	}
	defer userResp.Body.Close()
	userBody, _ := io.ReadAll(userResp.Body)

	orderResp, err := http.Get("http://service-b:9002/orders")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "service-b unavailable"})
		return
	}
	defer orderResp.Body.Close()
	orderBody, _ := io.ReadAll(orderResp.Body)

	c.JSON(http.StatusOK, gin.H{"users": string(userBody), "orders": string(orderBody)})
}

func healthCheck(c *gin.Context) {
	_ = os.Getenv("GATEWAY_ENV")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}