package event

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/estimate"
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

type HoldTimeEstimateListener struct {
	estimateRepository estimate.Repository
	prevEv             *NewHolderEvent
}

func NewHoldTimeEstimateListener(estimateRepository estimate.Repository) *HoldTimeEstimateListener {
	return &HoldTimeEstimateListener{estimateRepository: estimateRepository}
}

func (l *HoldTimeEstimateListener) Fire(ev NewHolderEvent) {
	if l.prevEv != nil && ev.AuthorUserId == l.prevEv.CurrentHolderUserId {
		log.Printf("calculating estimate prev=%v, curr=%v", l.prevEv, ev)
		err := l.estimateRepository.Save(ev.ts.Sub(l.prevEv.ts))
		if err != nil {
			log.Printf("can't save estimate: %s", err)
		}
	}
	l.prevEv = &ev
}
