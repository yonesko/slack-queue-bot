package listener

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"log"
)

type NotifySecondEventListener struct {
	slackApi *slack.Client
}

func NewNotifySecondEventListener(slackApi *slack.Client) *NotifySecondEventListener {
	return &NotifySecondEventListener{slackApi: slackApi}
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
