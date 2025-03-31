package commands

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"ninja-agent/bot/data"
	"ninja-agent/bot/utils"
)

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

	filter := bson.M{"date": utils.DateOnly(time.Now())}
	var result struct {
		ID    any         `bson:"_id"`
		Tasks []data.Task `bson:"tasks"`
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
