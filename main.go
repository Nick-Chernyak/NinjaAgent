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
)

func main() {

	fmt.Println("Starting bot...")
	fmt.Println("Reading environment variables...")

	mongoURI := os.Getenv("MONGO_URI")
	tgToken := os.Getenv("TG_TOKEN")
	allowedUserStr := os.Getenv("ALLOWED_USER")

	coll := initMongoAndgetcoll(mongoURI)
	allowedUser := setAllowedUser(allowedUserStr)
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	StartDayWatcher(context.Background(), coll, bot, allowedUser)

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

		switch cmd {
		case "todo":
			err = Todo(coll, bot, chatID, context.Background(), args)
			if err != nil {
				log.Println("Error adding todo:", err)
			}
			postShow(coll, bot, chatID, context.Background())

		case "show":
			err = Show(coll, bot, chatID, context.Background())
			if err != nil {
				log.Println("Error showing todos:", err)
			}
		case "done":
			err = Done(coll, bot, chatID, context.Background(), args)
			if err != nil {
				log.Println("Error marking todo as done:", err)
			}
			postShow(coll, bot, chatID, context.Background())
		case "remove":
			err = Remove(coll, bot, chatID, context.Background(), args)
			if err != nil {
				log.Println("Error removing todo:", err)
			}
			postShow(coll, bot, chatID, context.Background())
		}
	}
}

func postShow(coll *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64, ctx context.Context) {
	err := Show(coll, bot, chatID, context.Background())
	if err != nil {
		log.Println("Error showing todos:", err)
	}
}

func initMongoAndgetcoll(mongoURI string) *mongo.Collection {
	fmt.Println("Connecting to MongoDB...")
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	return client.Database("ninja_agent").Collection("todos")
}

func setAllowedUser(allowedUsersStr string) int64 {
	allowedUser, err := strconv.ParseInt(allowedUsersStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid ALLOWED_USER value: %v", err)
	}

	return allowedUser
}
