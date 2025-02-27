package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Définition de la structure Task
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// Stockage des tâches en mémoire
var tasks []Task
var nextID = 1

func main() {
	r := gin.Default()

	// Endpoint GET pour récupérer les tâches
	r.GET("/tasks", getTasks)

	// Endpoint POST pour ajouter une nouvelle tâche
	r.POST("/tasks", createTask)

	// Lancer le serveur
	r.Run(":8080")
}

// Récupérer toutes les tâches
func getTasks(c *gin.Context) {
	c.JSON(http.StatusOK, tasks)
}

// Ajouter une nouvelle tâche
func createTask(c *gin.Context) {
	var newTask Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON invalide"})
		return
	}
	newTask.ID = nextID
	nextID++
	tasks = append(tasks, newTask)
	c.JSON(http.StatusCreated, newTask)
}
