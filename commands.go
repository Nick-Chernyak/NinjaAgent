package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Todo(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, desc string) (err error) {

	desc = strings.TrimSpace(desc)
	if desc == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи описание задачи после команды."))
		return
	}

	task := Task{
		CreatedAt:   time.Now(),
		Description: desc,
		IsDone:      false,
	}

	filter := bson.M{"date": DateOnly(time.Now())}
	update := buildupdate(task)
	_, err = col.UpdateOne(
		ctx,
		filter,
		update)

	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при добавлении в базу."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "✅ Added!"))

	return
}

func Show(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context) (err error) {
	filter := bson.M{"date": DateOnly(time.Now())}
	var dayDoc Day
	err = col.FindOne(ctx, filter).Decode(&dayDoc)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при получении задач."))
		return
	}

	if len(dayDoc.Tasks) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Нет задач на сегодня."))
		return
	}

	tasksStr := FormatTasks(dayDoc.Tasks)
	bot.Send(tgbotapi.NewMessage(chatID, tasksStr))

	return
}

func Done(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, args string) (err error) {
	args = strings.TrimSpace(args)
	if args == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи номер задачи после команды."))
		return
	}

	taskNum, err := strconv.Atoi(args)
	if err != nil || taskNum < 1 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный номер задачи."))
		return
	}

	filter := bson.M{"date": DateOnly(time.Now())}
	taskIndex := taskNum - 1
	update := bson.M{
		"$set": bson.M{
			fmt.Sprintf("tasks.%d.is_done", taskIndex): true,
			fmt.Sprintf("tasks.%d.done_at", taskIndex): time.Now(),
		},
	}

	result, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при обновлении задачи."))
		return
	}

	if result.ModifiedCount == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "✅ Задача выполнена!"))

	return
}

func Remove(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, args string) (err error) {
	args = strings.TrimSpace(args)
	if args == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи номер задачи после команды."))
		return nil
	}

	taskNum, err := strconv.Atoi(args)
	if err != nil || taskNum < 1 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный номер задачи."))
		return nil
	}
	taskIndex := taskNum - 1

	filter := bson.M{"date": DateOnly(time.Now())}
	var result struct {
		ID    any    `bson:"_id"`
		Tasks []Task `bson:"tasks"`
	}

	err = col.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось найти задачи."))
		return nil
	}

	if taskIndex < 0 || taskIndex >= len(result.Tasks) {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный номер задачи."))
		return nil
	}

	result.Tasks = slices.Delete(result.Tasks, taskIndex, taskIndex+1)

	update := bson.M{"$set": bson.M{"tasks": result.Tasks}}
	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось удалить задачу."))
		return nil
	}

	bot.Send(tgbotapi.NewMessage(chatID, "🗑 Задача удалена!"))
	return nil
}

func buildupdate(task Task) bson.M {
	return bson.M{
		"$addToSet": bson.M{
			"tasks": task,
		},
	}
}
