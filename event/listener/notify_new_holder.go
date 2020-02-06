package listener

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	//"github.com/yonesko/slack-queue-bot/usecase"
	"log"
	"time"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent model.NewHolderEvent)
}

type NotifyNewHolderEventListener struct {
	slackApi *slack.Client
	//queueService usecase.QueueService
}

const waitForAck = time.Minute * 7

func NewNotifyNewHolderEventListener(slackApi *slack.Client) *NotifyNewHolderEventListener {
	return &NotifyNewHolderEventListener{slackApi: slackApi}
}

func (n *NotifyNewHolderEventListener) Fire(newHolderEvent model.NewHolderEvent) {
	//update holder : time and isSleeping
	//if holder is empty time is 0 isSleeping=false and return
	//on suc send msg: we wait for yor ack
	//wait for ack and pass then

	n.sendMsg(newHolderEvent.CurrentHolderUserId, fmt.Sprintf(i18n.P.MustGetString("your_turn_came"), waitForAck))

	time.AfterFunc(waitForAck, func() {
		n.passSleepingHolder(newHolderEvent.CurrentHolderUserId)
	})
}

func (n *NotifyNewHolderEventListener) passSleepingHolder(holderUserId string) {
	//err := n.queueService.Pass(holderUserId)
	//if err == usecase.YouAreNotHolder {
	//	return
	//}
	//if err == usecase.NoOneToPass {
	//	n.sendMsg(holderUserId, "я бы передал твой ход следующему, пока ты спишь, но ты один в очереди")
	//	return
	//}
	//if err != nil {
	//	log.Printf("can't passSleepingHolder %s", err)
	//	return
	//}
	//n.sendMsg(holderUserId, "твой ход передался следующему, пока ты спал")
}

func (n *NotifyNewHolderEventListener) sendMsg(holderUserId, txt string) {
	if holderUserId == "" {
		log.Printf("sendMsg user id is empty")
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
