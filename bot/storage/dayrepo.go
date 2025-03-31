package storage

import (
	"context"
	"fmt"
	"ninja-agent/bot/data"
	"ninja-agent/bot/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DayTasksRepo struct {
	collection *mongo.Collection
}

func NewDayTasksRepo(db *mongo.Database) *DayTasksRepo {
	return &DayTasksRepo{
		collection: db.Collection("todos"),
	}
}

var currentDayFilter bson.M = bson.M{"date": utils.DateOnly(time.Now())}

type TaskProj struct {
	ID    any         `bson:"_id"`
	Tasks []data.Task `bson:"tasks"`
}

func (repo DayTasksRepo) GetCurrentDay(ctx context.Context) (data.Day, error) {
	var result data.Day
	err := repo.collection.FindOne(ctx, currentDayFilter).Decode(&result)
	if err != nil {
		return data.Day{}, err
	}
	return result, nil
}

func (repo DayTasksRepo) GetCurrentTasks(ctx context.Context) ([]data.Task, error) {
	var result TaskProj
	err := repo.collection.FindOne(ctx, currentDayFilter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Tasks, nil
}

func (repo DayTasksRepo) AddTask(ctx context.Context, task data.Task) error {

	update := bson.M{
		"$addToSet": bson.M{
			"tasks": task,
		},
	}
	_, err := repo.collection.UpdateOne(
		ctx,
		currentDayFilter,
		update)

	return err
}

func (repo DayTasksRepo) RemoveTask(ctx context.Context, taskIndex int) error {

	var result TaskProj

	err := repo.collection.FindOne(ctx, currentDayFilter).Decode(&result)
	if err != nil {
		return err
	}

	result.Tasks = append(result.Tasks[:taskIndex], result.Tasks[taskIndex+1:]...)

	update := bson.M{"$set": bson.M{"tasks": result.Tasks}}
	_, err = repo.collection.UpdateOne(ctx, currentDayFilter, update)
	return err
}

func (repo DayTasksRepo) MarkTaskAsDone(ctx context.Context, taskIndex int) error {
	update := bson.M{
		"$set": bson.M{
			fmt.Sprintf("tasks.%d.is_done", taskIndex): true,
			fmt.Sprintf("tasks.%d.done_at", taskIndex): time.Now(),
		},
	}

	result, err := repo.collection.UpdateOne(ctx, currentDayFilter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no task found at index %d", taskIndex)
	}

	return nil
}
