package providers

import (
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
)

// CreateTelegramService creates a Telegram notification service
func CreateTelegramService(botToken string, chatIDs ...int64) (Notifier, error) {
	telegramService, err := telegram.New(botToken)
	if err != nil {
		return nil, err
	}

	// Add chat IDs
	for _, chatID := range chatIDs {
		telegramService.AddReceivers(chatID)
	}

	notifier := notify.New()
	notifier.UseServices(telegramService)

	return NewService(notifier), nil
}
