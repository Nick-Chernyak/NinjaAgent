package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	allowedUserIds := os.Getenv("ALLOWED_USER")
	if allowedUserIds == "" {
		fmt.Println("ALLOWED_USER is not set")
		return
	}

	parts := strings.Split(allowedUserIds, ",")
	allowedUsers := make([]int64, 0, len(parts))
	mongoClient := initMongo(mongoURI)
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	background.StartDayWatcher(context.Background(), mongoClient.Database("ninja_agent").Collection("todos"), bot, allowedUsers)

	executor := NewCommandExecutor(mongoClient, bot)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		if containsId(allowedUsers, int64(update.Message.From.ID)) == false {
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

func containsId(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
