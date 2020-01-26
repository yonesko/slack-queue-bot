package event

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/estimate"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/user"
	"log"
	"time"
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

type HoldTimeEstimateListener struct {
	estimateRepository estimate.Repository
	prevEv             *NewHolderEvent
}

func NewHoldTimeEstimateListener(estimateRepository estimate.Repository) *HoldTimeEstimateListener {
	return &HoldTimeEstimateListener{estimateRepository: estimateRepository}
}

func (l *HoldTimeEstimateListener) Fire(ev NewHolderEvent) {
	if l.prevEv != nil && ev.AuthorUserId == ev.PrevHolderUserId {
		log.Printf("calculating estimate prev=%#v, curr=%#v", l.prevEv, ev)
		l.calcEstimate(ev.Ts.Sub(l.prevEv.Ts))
	}
	l.prevEv = &ev
}

func (l *HoldTimeEstimateListener) calcEstimate(duration time.Duration) {
	estimate, err := l.estimateRepository.Get()
	if err != nil {
		log.Printf("can't calc estimate: %s", err)
		return
	}
	err = l.estimateRepository.Save(estimate.AddOne(duration))
	if err != nil {
		log.Printf("can't calc estimate: %s", err)
		return
	}
}
