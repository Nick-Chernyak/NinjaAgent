package background

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"ninja-agent/bot/data"
	"ninja-agent/bot/utils"
)

func StartDayWatcher(ctx context.Context, col *mongo.Collection, bot *tgbotapi.BotAPI, allowedUsers []int64) {

	fmt.Println("🕒 Запуск DayWatcher...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("📴 DayWatcher остановлен")
				return
			default:
				for _, chatID := range allowedUsers {
					err := ensureDay(col, bot, chatID)

					if err != nil {
						log.Println("DayWatcher error:", err)
					}
					time.Sleep(1 * time.Hour)
				}
			}
		}
	}()
}

func ensureDay(col *mongo.Collection, bot *tgbotapi.BotAPI, chatID int64) error {
	today := utils.DateOnly(time.Now())

	count, err := col.CountDocuments(context.Background(), currentdayAndchatFilter(chatID))
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	day := data.Day{
		Date:   today,
		Tasks:  generateRecurringTasks(),
		ChatID: chatID,
	}

	_, err = col.InsertOne(context.Background(), day)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "📅 Новый день создан!\nRecurring задачи добавлены.")
	_, _ = bot.Send(msg)

	return nil
}

func generateRecurringTasks() []data.Task {
	return []data.Task{
		{
			ID:          primitive.NewObjectID(),
			Description: "📓 Написать 3 мысли",
			CreatedAt:   time.Now(),
			IsDone:      false,
		},
	}
}

func currentdayAndchatFilter(chatID int64) bson.M {
	return bson.M{
		"date":    utils.DateOnly(time.Now()),
		"chat_id": chatID,
	}
}
