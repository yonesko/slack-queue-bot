package listener

import (
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
)

type NewSecondEventListener interface {
	Fire(newHolderEvent model.NewSecondEvent)
}

type NotifyNewSecondEventListener struct {
	gateway gateway.Gateway
}

func NewNotifyNewSecondEventListener(gateway gateway.Gateway) *NotifyNewSecondEventListener {
	return &NotifyNewSecondEventListener{gateway: gateway}
}

func (n *NotifyNewSecondEventListener) Fire(ev model.NewSecondEvent) {
	n.gateway.SendAndLog(ev.CurrentSecondUserId, i18n.L.MustGet("you_are_the_second"))
}
