package commands

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"ninja-agent/bot/data"
	"ninja-agent/bot/storage"
)

func Show(rep *storage.DayTasksRepo, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context) (err error) {

	tasks, err := rep.GetCurrentTasks(ctx, chatID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при получении задач."))
		return
	}

	if len(tasks) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Нет задач на сегодня."))
		return
	}

	tasksStr := FormatTasks(tasks)
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
