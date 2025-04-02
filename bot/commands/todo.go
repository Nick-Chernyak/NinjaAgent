package commands

import (
	"context"
	"strings"
	"time"

	"ninja-agent/bot/data"
	"ninja-agent/bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Todo(rep *storage.DayTasksRepo, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, desc string) (err error) {

	desc = strings.TrimSpace(desc)
	if desc == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи описание задачи после команды."))
		return
	}

	task := data.Task{
		CreatedAt:   time.Now(),
		Description: desc,
		IsDone:      false,
	}

	err = rep.AddTask(ctx, chatID, task)

	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при добавлении в базу."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "✅ Added!"))

	Show(rep, bot, chatID, ctx)

	return
}
