package main

import (
	"encoding/json" // Pour manipuler les fichiers JSON
	"fmt"           // Pour afficher les erreurs
	"net/http"      // Pour gérer les requêtes HTTP
	"os"            // Pour lire et écrire dans des fichiers
	"strconv"       // Pour convertir les types
	"sync"          // Pour éviter les conflits d'accès concurrentiel
	"time"

	"github.com/gin-gonic/gin" // Framework Gin pour gérer les routes HTTP
)

// Définition de la structure Task
type Task struct {
	ID    int    `json:"id"`    // Identifiant unique de la tâche, sérialisé en JSON sous "id"
	Title string `json:"title"` // Titre de la tâche, sérialisé sous "title"
}

// Variables globales
var (
	tasks  []Task     // Liste des tâches stockées en mémoire
	mutex  sync.Mutex // Mutex pour gérer l'accès concurrentiel aux tâches
	nextID = 1        // ID incrémental pour attribuer un identifiant unique aux nouvelles tâches
)

// Nom du fichier JSON où seront stockées les tâches
const taskFile = "tasks.json"

// Fonction pour sauvegarder les tâches dans le fichier JSON (sans ioutil)
func saveTasksToFile() {
	data, err := json.MarshalIndent(tasks, "", "  ") // Convertir la liste des tâches en JSON formaté
	if err != nil {
		fmt.Println("Erreur d'encodage JSON :", err)
		return
	}

	// Écrire les données JSON dans le fichier "tasks.json"
	err = os.WriteFile(taskFile, data, 0644)
	if err != nil {
		fmt.Println("Erreur d'écriture dans le fichier :", err)
	}
}

// Fonction pour charger les tâches depuis le fichier JSON au démarrage (sans ioutil)
func loadTasksFromFile() {
	// Vérifier si le fichier tasks.json existe
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		return // Si le fichier n'existe pas encore, on ne fait rien
	}

	// Lire le contenu du fichier JSON
	data, err := os.ReadFile(taskFile)
	if err != nil {
		fmt.Println("Erreur de lecture du fichier :", err)
		return
	}

	// Décoder le JSON en une liste de tâches
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		fmt.Println("Erreur de décodage JSON :", err)
		return
	}

	// Mettre à jour l'ID suivant en fonction des tâches existantes
	for _, task := range tasks {
		if task.ID >= nextID {
			nextID = task.ID + 1 // S'assurer que le prochain ID est unique
		}
	}
}

func main() {
	// Charger les tâches existantes au démarrage du serveur
	loadTasksFromFile()

	// Créer un routeur Gin
	r := gin.Default()

	// Route GET /tasks pour récupérer la liste des tâches
	r.GET("/tasks", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		c.JSON(http.StatusOK, tasks)
	})

	// Route POST /tasks pour ajouter une nouvelle tâche
	r.POST("/tasks", func(c *gin.Context) {
		var newTask Task // Déclaration d'une nouvelle tâche

		// Vérifier que le JSON envoyé est valide
		if err := c.ShouldBindJSON(&newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"}) // Retourner une erreur 400 si le JSON est incorrect
			return
		}

		mutex.Lock()
		newTask.ID = nextID // Assigner un ID unique
		nextID++            // Incrémenter l'ID pour la prochaine tâche
		tasks = append(tasks, newTask)
		saveTasksToFile() // Sauvegarder dans le fichier JSON
		mutex.Unlock()

		c.JSON(http.StatusCreated, newTask) // Retourner la tâche créée avec un code 201 Created
	})

	// Route PUT /tasks/:id pour modifier une tâche existante
	r.PUT("/tasks/:id", func(c *gin.Context) {
		idParam := c.Param("id")             // Récupérer l'ID passé en paramètre dans l'URL
		taskID, err := strconv.Atoi(idParam) // Convertir l'ID en entier
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"}) // Retourner une erreur 400 si l'ID est invalide
			return
		}

		var updatedTask Task // Déclaration de la tâche mise à jour
		if err := c.ShouldBindJSON(&updatedTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"}) // Retourner une erreur 400 si le JSON est invalide
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		// Parcourir la liste des tâches pour trouver celle à modifier
		for i, task := range tasks {
			if task.ID == taskID { // Si l'ID correspond
				tasks[i].Title = updatedTask.Title // Modifier le titre de la tâche
				saveTasksToFile()                  // Sauvegarder les changements
				c.JSON(http.StatusOK, tasks[i])    // Retourner la tâche mise à jour
				return
			}
		}

		// Retourner une erreur 404 si l'ID de la tâche n'est pas trouvé
		c.JSON(http.StatusNotFound, gin.H{"error": "Tâche non trouvée"})
	})

	// Route DELETE /tasks/:id pour supprimer une tâche
	r.DELETE("/tasks/:id", func(c *gin.Context) {
		idParam := c.Param("id")             // Récupérer l'ID passé en paramètre
		taskID, err := strconv.Atoi(idParam) // Convertir l'ID en entier
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"}) // Retourner une erreur 400 si l'ID est invalide
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		// Parcourir la liste des tâches pour trouver celle à supprimer
		for i, task := range tasks {
			if task.ID == taskID { // Si l'ID correspond
				tasks = append(tasks[:i], tasks[i+1:]...) // Supprimer la tâche de la liste
				saveTasksToFile()                         // Sauvegarder les changements
				c.JSON(http.StatusOK, gin.H{"message": "Tâche supprimée"}) // Confirmer la suppression
				return
			}
		}

		// Retourner une erreur 404 si la tâche n'existe pas
		c.JSON(http.StatusNotFound, gin.H{"error": "Tâche non trouvée"})
	})

	// Route /tasks/process pour exécuter une tâche en arrière-plan
	r.GET("/tasks/process", func(c *gin.Context) {
		go func() {
			fmt.Println("Démarrage du traitement en arrière-plan...")
			time.Sleep(5 * time.Second) // Simule un traitement long de 5 secondes
			fmt.Println("Traitement terminé.")
		}()

		// Répondre immédiatement sans attendre la fin du traitement
		c.JSON(http.StatusAccepted, gin.H{"message": "Traitement lancé en arrière-plan"})
	})

	// Lancer le serveur sur le port 8080
	r.Run(":8080")
}
