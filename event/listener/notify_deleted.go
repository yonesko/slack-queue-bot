package listener

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"log"
)

type DeletedEventListener interface {
	Fire(newHolderEvent model.DeletedEvent)
}

type notifyDeletedEventListener struct {
	slackApi *slack.Client
	gateway  gateway.Gateway
}

func NewNotifyDeletedEventListener(slackApi *slack.Client) *notifyDeletedEventListener {
	return &notifyDeletedEventListener{slackApi: slackApi}
}

func (n *notifyDeletedEventListener) Fire(ev model.DeletedEvent) {
	_, _, err := n.slackApi.PostMessage(ev.CurrentSecondUserId,
		slack.MsgOptionText(i18n.L.MustGet("you_are_the_second"), true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
	}
}
