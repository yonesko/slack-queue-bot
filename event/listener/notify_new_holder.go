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

const waitForAck = time.Minute * 7

func NewNotifyNewHolderEventListener(slackApi *slack.Client) *NotifyNewHolderEventListener {
	return &NotifyNewHolderEventListener{slackApi: slackApi}
}

func (n *NotifyNewHolderEventListener) Fire(newHolderEvent model.NewHolderEvent) {
	curHolder := newHolderEvent.CurrentHolderUserId
	err := n.queueService.UpdateOnNewHolder()
	if err != nil {
		log.Printf("can't UpdateOnNewHolder, return")
		return
	}
	if curHolder == "" {
		log.Printf("holder is empty, return")
		return
	}

	err = n.sendMsg(curHolder, fmt.Sprintf(i18n.P.MustGetString("your_turn_came"), waitForAck))
	if err != nil {
		log.Printf("can't notify holder %s: %s", curHolder, err)
		return
	}

	time.AfterFunc(waitForAck, func() { n.passSleepingHolder(curHolder) })
}

func (n *NotifyNewHolderEventListener) passSleepingHolder(holderUserId string) {
	err := n.queueService.Pass(holderUserId)
	if err == usecase.YouAreNotHolder {
		log.Printf("passSleepingHolder %s", err)
		return
	}
	if err == usecase.NoOneToPass {
		n.sendMsgAndLog(holderUserId, "я бы передал твой ход следующему, пока ты спишь, но ты один в очереди")
		return
	}
	if err != nil {
		log.Printf("can't passSleepingHolder %s", err)
		return
	}
	n.sendMsgAndLog(holderUserId, "твой ход передался следующему, пока ты спал")
}

func (n *NotifyNewHolderEventListener) sendMsg(userId, txt string) error {
	if userId == "" {
		log.Printf("sendMsg user id is empty")
		return nil
	}
	_, _, err := n.slackApi.PostMessage(userId,
		slack.MsgOptionText(txt, true),
		slack.MsgOptionAsUser(true),
	)
	return err
}

func (n *NotifyNewHolderEventListener) sendMsgAndLog(userId, txt string) {
	err := n.sendMsg(userId, txt)
	if err != nil {
		log.Printf("can't send %s %s", userId, txt)
	}
}
