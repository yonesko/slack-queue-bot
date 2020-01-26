package event

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/user"
	"log"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent NewHolderEvent)
}

type NotifyNewHolderEventListener struct {
	slackApi       *slack.Client
	userRepository user.Repository
}

func NewNotifyNewHolderEventListener(slackApi *slack.Client, userRepository user.Repository) *NotifyNewHolderEventListener {
	return &NotifyNewHolderEventListener{slackApi: slackApi, userRepository: userRepository}
}

func (n *NotifyNewHolderEventListener) Fire(newHolderEvent NewHolderEvent) {
	_, _, err := n.slackApi.PostMessage(newHolderEvent.CurrentHolderUserId,
		slack.MsgOptionText(i18n.P.MustGetString("your_turn_came"), true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
	}
}
