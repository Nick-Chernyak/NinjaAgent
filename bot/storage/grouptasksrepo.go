package storage

import (
	"context"
	"fmt"
	"log"
	"ninja-agent/bot/data"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GroupTaskRepo struct {
	collection *mongo.Collection
}

func NewGroupTaskRepo(db *mongo.Database) *GroupTaskRepo {
	return &GroupTaskRepo{
		collection: db.Collection("group_tasks"),
	}
}

func (repo GroupTaskRepo) GetCurrentTasks(ctx context.Context, chatID int64) ([]data.GroupTask, error) {
	var result []data.GroupTask
	allNotDoneTasksFilter := bson.M{"is_done": false, "is_archived": false}
	cur, err := repo.collection.Find(ctx, allNotDoneTasksFilter)
	if err != nil {
		log.Println("Error finding tasks:", err)
		return nil, err
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo GroupTaskRepo) AddTask(ctx context.Context, chatID int64, task data.GroupTask) error {
	_, err := repo.collection.InsertOne(ctx, task)
	if err != nil {
		log.Println("Error inserting task:", err)
	}
	return err
}

func (repo GroupTaskRepo) RemoveTask(ctx context.Context, chatID int64, taskIndex int) error {

	ct, err := repo.GetCurrentTasks(ctx, chatID)
	if err != nil {
		log.Println("Error getting tasks:", err)
		return err
	}

	if taskIndex < 0 || taskIndex >= len(ct) {
		return fmt.Errorf("task index out of range: %d", taskIndex)
	}

	removeTaskID := bson.M{"_id": ct[taskIndex].ID}
	_, err = repo.collection.DeleteOne(ctx, removeTaskID)
	if err != nil {
		log.Println("Error deleting task:", err)
	}
	return err
}

func (repo GroupTaskRepo) MarkTaskAsDone(ctx context.Context, chatID int64, taskIndex int) error {

	ct, err := repo.GetCurrentTasks(ctx, chatID)
	if err != nil {
		log.Println("Error getting tasks:", err)
		return err
	}

	if taskIndex < 0 || taskIndex >= len(ct) {
		return fmt.Errorf("task index out of range: %d", taskIndex)
	}

	filter := bson.M{"_id": ct[taskIndex].ID}
	update := bson.M{"$set": bson.M{"is_done": true}}
	_, err = repo.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Error updating task:", err)
	}
	return err
}
