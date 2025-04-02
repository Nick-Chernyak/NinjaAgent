package commands

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"ninja-agent/bot/storage"
)

func Done(rep *storage.DayTasksRepo, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, args string) (err error) {
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

	err = rep.MarkTaskAsDone(ctx, chatID, taskNum-1)

	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при обновлении задачи."))
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "✅ Задача выполнена!"))
	Show(rep, bot, chatID, ctx)

	return
}
