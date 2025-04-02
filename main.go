package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ninja-agent/bot/background"
	"slices"
)

func main() {

	fmt.Println("Starting bot...")
	fmt.Println("Reading environment variables...")

	mongoURI := os.Getenv("MONGO_URI")
	tgToken := os.Getenv("TG_TOKEN")
	allowedUsers, err := parseInt64ListFromEnv("ALLOWED_USERS")
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
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

		if handler, ok := executor.handlers[cmd]; ok {
			handler(chatID, context.Background(), args)
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
	return slices.Contains(slice, val)
}

func parseInt64ListFromEnv(envVar string) ([]int64, error) {
	raw := os.Getenv(envVar)
	if raw == "" {
		return nil, fmt.Errorf("env var %q is not set", envVar)
	}

	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))

	for _, p := range parts {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64 in %q: %v", envVar, err)
		}
		result = append(result, n)
	}

	return result, nil
}
