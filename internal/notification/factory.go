package notification

import (
	c "dns-monitor/internal/common"
	"dns-monitor/internal/notification/providers"
	"fmt"
	"strings"
)

// Factory provides methods to create different notification services
type Factory struct {
	config *c.NotificationConfig
}

// NewFactory creates a new notification factory with the given configuration
func NewFactory(cfg *c.NotificationConfig) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateNotifier creates a notifier based on the configuration
func (f *Factory) CreateNotifier() (providers.Notifier, error) {
	switch strings.ToLower(f.config.NotifierType) {
	case c.NotifierTypePushover:
		return f.createPushoverService()
	case c.NotifierTypeTelegram:
		return f.createTelegramService()
	// Add more cases for other notifier types as needed
	default:
		return nil, fmt.Errorf("unsupported notifier type: %s", f.config.NotifierType)
	}
}

// createPushoverService creates a Pushover notification service from environment variables
func (f *Factory) createPushoverService() (providers.Notifier, error) {
	return providers.CreatePushoverService(f.config.PushoverUser, f.config.PushoverToken)
}

// createTelegramService creates a Telegram notification service from environment variables
func (f *Factory) createTelegramService() (providers.Notifier, error) {
	return providers.CreateTelegramService(f.config.TelegramBotToken, f.config.TelegramChatIDs...)
}
