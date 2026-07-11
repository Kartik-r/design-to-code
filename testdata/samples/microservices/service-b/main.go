package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var db *sql.DB

func main() {
	dsn := os.Getenv("SERVICE_B_DB_DSN")
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.GET("/orders", listOrders)
	router.Run(":9002")
}

func listOrders(c *gin.Context) {
	rows, _ := db.Query("SELECT id, total FROM orders")
	defer rows.Close()
	c.JSON(http.StatusOK, gin.H{"orders": []string{}})
}