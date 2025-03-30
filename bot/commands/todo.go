package commands

import (
	"context"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	types "ninja-agent/bot"
	"ninja-agent/bot/utils"
)

func Todo(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, desc string) (err error) {

	desc = strings.TrimSpace(desc)
	if desc == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи описание задачи после команды."))
		return
	}

	task := types.Task{
		CreatedAt:   time.Now(),
		Description: desc,
		IsDone:      false,
	}

	filter := bson.M{"date": utils.DateOnly(time.Now())}
	update := bson.M{
		"$addToSet": bson.M{
			"tasks": task,
		},
	}
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
