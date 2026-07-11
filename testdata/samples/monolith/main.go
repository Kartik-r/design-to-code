package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")

	var err error
	db, err = sql.Open("postgres", buildDSN(dbHost, dbPort, dbUser))
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.GET("/users/:id", getUser)
	router.POST("/users", createUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)
	router.GET("/health", healthCheck)

	router.Run(":8080")
}

func buildDSN(host, port, user string) string {
	return "host=" + host + " port=" + port + " user=" + user
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT id, name FROM users WHERE id = $1", id)
	var name string
	row.Scan(&id, &name)
	c.JSON(http.StatusOK, gin.H{"id": id, "name": name})
}

func createUser(c *gin.Context) {
	db.Exec("INSERT INTO users (name) VALUES ($1)", c.PostForm("name"))
	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}

func updateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func deleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}