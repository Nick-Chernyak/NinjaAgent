package group

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"ninja-agent/data/storage"
)

func Remove(rep *storage.GroupTaskRepo, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context, args string) (err error) {
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

	err = rep.RemoveTask(ctx, chatID, taskIndex)

	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось удалить задачу."))
		return nil
	}

	bot.Send(tgbotapi.NewMessage(chatID, "🗑 Задача удалена!"))
	Show(rep, bot, chatID, ctx)
	return nil
}
