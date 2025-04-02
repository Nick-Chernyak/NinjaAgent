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
	allowedGroupChatId, err := strconv.ParseInt(os.Getenv("ALLOWED_GROUP_CHAT"), 10, 64)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
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

	background.StartDayWatcher(context.Background(), mongoClient.Database("ninja_agent"), bot, allowedUsers)

	dailyExecutor := NewDailyCommandExecutor(mongoClient, bot)
	groupExecutor := NewGroupExecutor(mongoClient, bot)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		if !slices.Contains(allowedUsers, int64(update.Message.From.ID)) {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ У вас нет доступа к этому боту."))
			continue
		}

		chatID := getchatID(update, allowedGroupChatId)
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		if isGroupChat(update) {
			if handler, ok := groupExecutor.handlers[cmd]; ok {
				handler(chatID, context.Background(), args)
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Неизвестная команда."))
			}
		}

		if handler, ok := dailyExecutor.handlers[cmd]; ok {
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

func getchatID(upd tgbotapi.Update, allowed int64) int64 {
	if upd.Message.Chat.ID == allowed && isGroupChat(upd) {
		return upd.Message.Chat.ID
	}
	return upd.Message.From.ID
}

func isGroupChat(upd tgbotapi.Update) bool {
	return upd.Message.Chat.Type == "group"
}
