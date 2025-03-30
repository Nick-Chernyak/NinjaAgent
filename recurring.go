package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartDayWatcher(ctx context.Context, col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64) {

	fmt.Println("🕒 Запуск DayWatcher...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("📴 DayWatcher остановлен")
				return
			default:
				err := ensureDay(col, bot, chatID)
				if err != nil {
					log.Println("DayWatcher error:", err)
				}
				time.Sleep(1 * time.Hour)
			}
		}
	}()
}

func ensureDay(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64) error {
	today := DateOnly(time.Now())

	filter := bson.M{"day": today}
	count, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	day := Day{
		Date:  today,
		Tasks: generateRecurringTasks(),
	}

	_, err = col.InsertOne(context.Background(), day)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "📅 Новый день создан!\nRecurring задачи добавлены.")
	_, _ = bot.Send(msg)

	return nil
}

func generateRecurringTasks() []Task {
	return []Task{
		{
			ID:          primitive.NewObjectID(),
			Description: "🌅 Утренняя медитация",
			CreatedAt:   time.Now(),
			IsDone:      false,
		},
		{
			ID:          primitive.NewObjectID(),
			Description: "📓 Написать 3 мысли",
			CreatedAt:   time.Now(),
			IsDone:      false,
		},
	}
}
