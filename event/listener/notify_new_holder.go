package listener

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	model2 "github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/user"
	"log"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent model2.NewHolderEvent)
}

type NotifyNewHolderEventListener struct {
	slackApi       *slack.Client
	userRepository user.Repository
}

func NewNotifyNewHolderEventListener(slackApi *slack.Client, userRepository user.Repository) *NotifyNewHolderEventListener {
	return &NotifyNewHolderEventListener{slackApi: slackApi, userRepository: userRepository}
}

func (n *NotifyNewHolderEventListener) Fire(newHolderEvent model2.NewHolderEvent) {
	if newHolderEvent.CurrentHolderUserId == "" {
		return
	}
	_, _, err := n.slackApi.PostMessage(newHolderEvent.CurrentHolderUserId,
		slack.MsgOptionText(i18n.P.MustGetString("your_turn_came"), true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
	}
}
