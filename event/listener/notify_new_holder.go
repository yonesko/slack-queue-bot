package listener

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"log"
	"time"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent model.NewHolderEvent)
}

type NotifyNewHolderEventListener struct {
	slackApi *slack.Client
}

const waitAckDur = time.Minute * 7

func NewNotifyNewHolderEventListener(slackApi *slack.Client) *NotifyNewHolderEventListener {
	return &NotifyNewHolderEventListener{slackApi: slackApi}
}

func (n *NotifyNewHolderEventListener) Fire(newHolderEvent model.NewHolderEvent) {
	if newHolderEvent.CurrentHolderUserId == "" {
		return
	}
	txt := fmt.Sprintf(i18n.P.MustGetString("your_turn_came"), waitAckDur)
	_, _, err := n.slackApi.PostMessage(newHolderEvent.CurrentHolderUserId,
		slack.MsgOptionText(txt, true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
		return
	}

}
