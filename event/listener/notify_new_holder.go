package listener

import (
	"github.com/yonesko/slack-queue-bot/model"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent model.NewHolderEvent)
}
