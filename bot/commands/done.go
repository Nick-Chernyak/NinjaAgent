package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"ninja-agent/bot/utils"
)

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

	filter := bson.M{"date": utils.DateOnly(time.Now())}
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
