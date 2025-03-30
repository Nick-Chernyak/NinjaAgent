package main

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Day struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Tasks []Task             `bson:"tasks"`
	Date  time.Time          `bson:"date"`
}

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	Description string             `bson:"description"`
	IsDone      bool               `bson:"is_done"`
	DoneAt      *time.Time         `bson:"done_at,omitempty"`
}

func FormatTasks(tasks []Task) string {
	var sb strings.Builder
	for i, task := range tasks {
		status := "❌"
		if task.IsDone {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, status, task.Description))
	}
	return sb.String()
}
