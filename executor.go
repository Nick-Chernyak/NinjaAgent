package main

import (
	"context"
	"ninja-agent/bot/commands"
	"ninja-agent/bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommandExecutor struct {
	handlers map[string]func(chatID int64, ctx context.Context, args string)
}

func NewCommandExecutor(client *mongo.Client, bot *tgbotapi.BotAPI) *CommandExecutor {
	executor := &CommandExecutor{
		handlers: make(map[string]func(chatID int64, ctx context.Context, args string)),
	}

	repo := storage.NewDayTasksRepo(client.Database("ninja_agent"))

	executor.handlers["todo"] = func(chatID int64, ctx context.Context, args string) {
		commands.Todo(repo, bot, chatID, ctx, args)
	}
	executor.handlers["done"] = func(chatID int64, ctx context.Context, args string) {
		commands.Done(repo, bot, chatID, ctx, args)
	}
	executor.handlers["remove"] = func(chatID int64, ctx context.Context, args string) {
		commands.Remove(repo, bot, chatID, ctx, args)
	}
	executor.handlers["show"] = func(chatID int64, ctx context.Context, args string) {
		commands.Show(repo, bot, chatID, ctx)
	}

	return executor
}
