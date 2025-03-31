package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ninja-agent/bot/background"
)

func main() {

	fmt.Println("Starting bot...")
	fmt.Println("Reading environment variables...")

	mongoURI := os.Getenv("MONGO_URI")
	tgToken := os.Getenv("TG_TOKEN")
	allowedUserStr := os.Getenv("ALLOWED_USER")

	allowedUser := setAllowedUser(allowedUserStr)
	mongoClient := initMongo(mongoURI)
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	background.StartDayWatcher(context.Background(), mongoClient.Database("ninja-agent").Collection("todos"), bot, allowedUser)

	executor := NewCommandExecutor(mongoClient, bot)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		if update.Message.From.ID != allowedUser {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ У вас нет доступа к этому боту."))
			continue
		}

		chatID := update.Message.Chat.ID
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		if hanlder, ok := executor.handlers[cmd]; ok {
			hanlder(chatID, context.Background(), args)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неизвестная команда."))
		}
	}
}

func initMongo(mongoURI string) *mongo.Client {
	fmt.Println("Connecting to MongoDB...")
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func setAllowedUser(allowedUsersStr string) int64 {
	allowedUser, err := strconv.ParseInt(allowedUsersStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid ALLOWED_USER value: %v", err)
	}

	return allowedUser
}
