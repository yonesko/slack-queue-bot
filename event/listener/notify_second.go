package listener

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/user"
	"log"
)

type NotifySecondEventListener struct {
	slackApi       *slack.Client
	userRepository user.Repository
}

func NewNotifySecondEventListener(slackApi *slack.Client, userRepository user.Repository) *NotifySecondEventListener {
	return &NotifySecondEventListener{slackApi: slackApi, userRepository: userRepository}
}

func (n *NotifySecondEventListener) Fire(newHolderEvent model.NewHolderEvent) {
	if newHolderEvent.SecondUserId == "" {
		return
	}
	_, _, err := n.slackApi.PostMessage(newHolderEvent.SecondUserId,
		slack.MsgOptionText(i18n.P.MustGetString("you_are_the_second"), true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
	}
}
