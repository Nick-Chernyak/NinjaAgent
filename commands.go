package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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

	filter := builddayfilter()
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
	filter := builddayfilter()
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

	filter := builddayfilter()
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
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи номер задачи для удаления."))
		return
	}

	taskNum, err := strconv.Atoi(args)
	if err != nil || taskNum < 1 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный номер задачи."))
		return
	}

	filter := builddayfilter()
	update := bson.M{
		"$pull": bson.M{
			"tasks": bson.M{"$eq": taskNum},
		},
	}

	result, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при удалении задачи."))
		return
	}

	if result.ModifiedCount == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "🗑️ Задача удалена."))

	return
}

func builddayfilter() bson.M {
	now := time.Now()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return bson.M{"day": day}
}

func buildupdate(task Task) bson.M {
	return bson.M{
		"$addToSet": bson.M{
			"tasks": task,
		},
	}
}
