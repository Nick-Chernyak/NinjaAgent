package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ToDo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	Description string             `bson:"description"`
	IsDone      bool               `bson:"is_done"`
	DoneAt      *time.Time         `bson:"done_at,omitempty"`
}

type CacheRequest struct {
	ChatID int64
	Action string
	Todo   *ToDo
	Reply  chan []ToDo
}

func StartCache(ctx context.Context, mongoColl *mongo.Collection, ttl time.Duration) chan<- CacheRequest {
	reqs := make(chan CacheRequest)
	type entry struct {
		Todos     []ToDo
		ExpiresAt time.Time
	}

	cache := make(map[int64]entry)

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case req := <-reqs:
				ent, found := cache[req.ChatID]
				if !found || time.Now().After(ent.ExpiresAt) {
					cursor, _ := mongoColl.Find(context.TODO(), bson.M{
						"created_at": bson.M{"$gte": time.Now().Truncate(24 * time.Hour)},
					})
					var todos []ToDo
					cursor.All(context.TODO(), &todos)
					ent = entry{Todos: todos, ExpiresAt: time.Now().Add(ttl)}
					cache[req.ChatID] = ent
				}
				if req.Action == "add" && req.Todo != nil {
					ent.Todos = append(ent.Todos, *req.Todo)
					cache[req.ChatID] = ent
				}
				req.Reply <- ent.Todos

			case <-ticker.C:
				for id, ent := range cache {
					if time.Now().After(ent.ExpiresAt) {
						delete(cache, id)
					}
				}
			}
		}
	}()

	return reqs
}

func main() {
	fmt.Println("Starting bot...")

	fmt.Println("Reading environment variables...")
	ctx := context.Background()
	mongoURI := os.Getenv("MONGO_URI")
	tgToken := os.Getenv("TG_TOKEN")
	allowedUsersStr := os.Getenv("ALLOWED_USER")
	allowedUsers, err := strconv.ParseInt(allowedUsersStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid ALLOWED_USER value: %v", err)
	}

	fmt.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	coll := client.Database("ninja_agent").Collection("todos")
	cache := StartCache(ctx, coll, 15*time.Minute)

	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		if update.Message.From.ID != allowedUsers {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ У вас нет доступа к этому боту."))
			continue
		}

		chatID := update.Message.Chat.ID
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		switch cmd {
		case "todo":
			desc := strings.TrimSpace(args)
			if desc == "" {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи описание задачи после команды."))
				continue
			}

			todo := &ToDo{
				CreatedAt:   time.Now(),
				Description: desc,
				IsDone:      false,
			}

			res, err := coll.InsertOne(ctx, todo)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при добавлении в базу."))
				continue
			}
			todo.ID = res.InsertedID.(primitive.ObjectID)

			reply := make(chan []ToDo)
			cache <- CacheRequest{ChatID: chatID, Action: "add", Todo: todo, Reply: reply}
			<-reply

			bot.Send(tgbotapi.NewMessage(chatID, "✅ Added!"))

		case "show":
			reply := make(chan []ToDo)
			cache <- CacheRequest{ChatID: chatID, Action: "show", Reply: reply}
			todos := <-reply
			msg := formatTodos(todos)
			bot.Send(tgbotapi.NewMessage(chatID, msg))

		case "complete":
			index, err := strconv.Atoi(strings.TrimSpace(args))
			if err != nil || index < 1 {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи номер задачи для завершения."))
				continue
			}
			reply := make(chan []ToDo)
			cache <- CacheRequest{ChatID: chatID, Action: "show", Reply: reply}
			todos := <-reply
			if index > len(todos) {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Такого номера нет."))
				continue
			}
			todo := todos[index-1]
			if todo.IsDone {
				bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Задача уже выполнена."))
				continue
			}
			todo.IsDone = true
			now := time.Now()
			todo.DoneAt = &now
			_, err = coll.UpdateByID(ctx, todo.ID, bson.M{"$set": bson.M{"is_done": true, "done_at": now}})
			if err == nil {
				bot.Send(tgbotapi.NewMessage(chatID, "✅ Задача выполнена."))
			}

		case "remove":
			index, err := strconv.Atoi(strings.TrimSpace(args))
			if err != nil || index < 1 {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Укажи номер задачи для удаления."))
				continue
			}
			reply := make(chan []ToDo)
			cache <- CacheRequest{ChatID: chatID, Action: "show", Reply: reply}
			todos := <-reply
			if index > len(todos) {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Такого номера нет."))
				continue
			}
			todo := todos[index-1]
			_, err = coll.DeleteOne(ctx, bson.M{"_id": todo.ID})
			if err == nil {
				bot.Send(tgbotapi.NewMessage(chatID, "🗑️ Задача удалена."))
			}

		case "edit":
			parts := strings.SplitN(args, " ", 2)
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Используй: /edit <номер> <новый текст>"))
				continue
			}
			index, err := strconv.Atoi(parts[0])
			if err != nil || index < 1 {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный номер задачи."))
				continue
			}
			newText := strings.TrimSpace(parts[1])
			if newText == "" {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Новый текст не должен быть пустым."))
				continue
			}
			reply := make(chan []ToDo)
			cache <- CacheRequest{ChatID: chatID, Action: "show", Reply: reply}
			todos := <-reply
			if index > len(todos) {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Такого номера нет."))
				continue
			}
			todo := todos[index-1]
			_, err = coll.UpdateByID(ctx, todo.ID, bson.M{"$set": bson.M{"description": newText}})
			if err == nil {
				bot.Send(tgbotapi.NewMessage(chatID, "✏️ Задача обновлена."))
			}
		}
	}
}

func formatTodos(todos []ToDo) string {
	if len(todos) == 0 {
		return "📭 На сегодня задач нет."
	}
	var sb strings.Builder
	for i, todo := range todos {
		status := "❌"
		if todo.IsDone {
			status = "✅"
		}
		sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, status, todo.Description))
	}
	return sb.String()
}
