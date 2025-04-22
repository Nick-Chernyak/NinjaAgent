package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"ninja-agent/data"
	"ninja-agent/data/storage"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RunAPI() {
	db := initMongo(os.Getenv("MONGO_URI"))
	repo := storage.NewDayTasksRepo(db)

	r := gin.Default()

	r.GET("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, getTasks(repo))
	})

	r.POST("/tasks", func(c *gin.Context) {
		var dto CreateTaskDto
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		task := data.Task{
			CreatedAt:   time.Now(),
			Description: dto.Description,
			IsDone:      false,
		}

		err := repo.AddTask(context.Background(), 456940399, task)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add task"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Task added successfully!"})
	})

	r.DELETE("/tasks", func(c *gin.Context) {
		var dto DeleteTaskDto
		if err := c.ShouldBindBodyWithJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := repo.RemoveTask(context.Background(), 456940399, dto.Number); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{"message": "Task removed successfully!"})
	})

	port := getPort()
	r.Run(":" + port)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func initMongo(mongoURI string) *mongo.Database {
	fmt.Println("Connecting to MongoDB...")
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	return client.Database("ninja_agent")
}

func getTasks(repo *storage.DayTasksRepo) []data.Task {
	tasks, err := repo.GetCurrentTasks(context.Background(), 456940399)
	if err != nil {
		log.Fatal(err)
	}
	return tasks
}

type CreateTaskDto struct {
	Description string `json:"description"`
}

type DeleteTaskDto struct {
	Number int `json:number`
}
