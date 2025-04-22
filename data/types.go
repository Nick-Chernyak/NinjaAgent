package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Day struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Tasks  []Task             `bson:"tasks"`
	Date   time.Time          `bson:"date"`
	ChatID int64              `bson:"chat_id"`
}

type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	Description string             `bson:"description"`
	IsDone      bool               `bson:"is_done"`
	DoneAt      *time.Time         `bson:"done_at,omitempty"`
}

type GroupTask struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	Description string             `bson:"description"`
	IsDone      bool               `bson:"is_done"`
	DoneAt      *time.Time         `bson:"done_at,omitempty"`
	IsArchived  bool               `bson:"is_archived"`
}
