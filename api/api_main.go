package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"ninja-agent/bot/data"
	"ninja-agent/bot/storage"
	"os"

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
