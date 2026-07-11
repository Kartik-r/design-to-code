package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var db *sql.DB

func main() {
	dsn := os.Getenv("SERVICE_A_DB_DSN")
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.GET("/users", listUsers)
	router.Run(":9001")
}

func listUsers(c *gin.Context) {
	rows, _ := db.Query("SELECT id, name FROM users")
	defer rows.Close()
	c.JSON(http.StatusOK, gin.H{"users": []string{}})
}