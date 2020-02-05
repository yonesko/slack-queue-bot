package listener

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/usecase"
	"log"
	"time"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent model.NewHolderEvent)
}

type NotifyNewHolderEventListener struct {
	slackApi     *slack.Client
	queueService usecase.QueueService
}

const waitAckDur = time.Minute * 7

func NewNotifyNewHolderEventListener(slackApi *slack.Client) *NotifyNewHolderEventListener {
	return &NotifyNewHolderEventListener{slackApi: slackApi}
}

func (n *NotifyNewHolderEventListener) Fire(newHolderEvent model.NewHolderEvent) {
	n.sendMsg(newHolderEvent.CurrentHolderUserId, fmt.Sprintf(i18n.P.MustGetString("your_turn_came"), waitAckDur))
}

func (n *NotifyNewHolderEventListener) passSleepingHolder(holderUserId string) {
	err := n.queueService.Pass(holderUserId)
	if err != nil {
		log.Printf("can't passSleepingHolder %s", err)
		return
	}
	n.sendMsg(holderUserId, "тебя выкинули пока ты спал, hasta la vista")
}

func (n *NotifyNewHolderEventListener) sendMsg(holderUserId, txt string) {
	if holderUserId == "" {
		return
	}
	_, _, err := n.slackApi.PostMessage(holderUserId,
		slack.MsgOptionText(txt, true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't sendMsg %s", err)
	}
}
