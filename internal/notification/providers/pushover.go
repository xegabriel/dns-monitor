package providers

import (
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/pushover"
)

func CreatePushoverService(userKey, appToken string) (Notifier, error) {
	notifier := notify.New()
	pushoverService := pushover.New(appToken)
	pushoverService.AddReceivers(userKey)
	notifier.UseServices(pushoverService)
	return NewService(notifier), nil
}
