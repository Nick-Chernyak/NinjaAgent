package main

import (
	"context"
	grcom "ninja-agent/bot/commands/group"
	"ninja-agent/bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type GroupExecutor struct {
	handlers map[string]func(chatID int64, ctx context.Context, args string)
}

func (exec GroupExecutor) GetHandler(command string) (func(chatID int64, ctx context.Context, args string), bool) {
	handler, ok := exec.handlers[command]
	return handler, ok
}

func NewGroupExecutor(client *mongo.Client, bot *tgbotapi.BotAPI) *GroupExecutor {
	executor := &GroupExecutor{
		handlers: make(map[string]func(chatID int64, ctx context.Context, args string)),
	}

	repo := storage.NewGroupTaskRepo(client.Database("ninja_agent"))

	executor.handlers["todo"] = func(chatID int64, ctx context.Context, args string) {
		grcom.Todo(repo, bot, chatID, ctx, args)
	}
	executor.handlers["done"] = func(chatID int64, ctx context.Context, args string) {
		grcom.Done(repo, bot, chatID, ctx, args)
	}
	executor.handlers["remove"] = func(chatID int64, ctx context.Context, args string) {
		grcom.Remove(repo, bot, chatID, ctx, args)
	}
	executor.handlers["show"] = func(chatID int64, ctx context.Context, args string) {
		grcom.Show(repo, bot, chatID, ctx)
	}

	return executor
}
