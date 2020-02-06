package listener

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"log"
)

type NewSecondEventListener interface {
	Fire(newHolderEvent model.NewSecondEvent)
}

type NotifyNewSecondEventListener struct {
	slackApi *slack.Client
}

func NewNotifyNewSecondEventListener(slackApi *slack.Client) *NotifyNewSecondEventListener {
	return &NotifyNewSecondEventListener{slackApi: slackApi}
}

func (n *NotifyNewSecondEventListener) Fire(ev model.NewSecondEvent) {
	if ev.CurrentSecondUserId == "" {
		return
	}
	_, _, err := n.slackApi.PostMessage(ev.CurrentSecondUserId,
		slack.MsgOptionText(i18n.L.MustGet("you_are_the_second"), true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
	}
}
