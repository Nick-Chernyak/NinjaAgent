package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"ninja-agent/bot/data"
	"ninja-agent/bot/utils"
)

func Show(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context) (err error) {
	filter := bson.M{"date": utils.DateOnly(time.Now())}
	var dayDoc data.Day
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

func FormatTasks(tasks []data.Task) string {
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
