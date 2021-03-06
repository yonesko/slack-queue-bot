package listener

import (
	"github.com/yonesko/slack-queue-bot/estimate"
	"github.com/yonesko/slack-queue-bot/model"
	"log"
	"time"
)

type NewHolderEventListener interface {
	Fire(newHolderEvent model.NewHolderEvent)
}
type HoldTimeEstimateListener struct {
	estimateRepository estimate.Repository
	prevEv             *model.NewHolderEvent
}

func NewHoldTimeEstimateListener(estimateRepository estimate.Repository) *HoldTimeEstimateListener {
	return &HoldTimeEstimateListener{estimateRepository: estimateRepository}
}
func (l *HoldTimeEstimateListener) Fire(ev model.NewHolderEvent) {
	if l.prevEv != nil && ev.AuthorUserId == ev.PrevHolderUserId {
		duration := ev.Ts.Sub(l.prevEv.Ts)
		if isTimeSeemsLegit(duration) {
			log.Printf("hold time was %s", duration.String())
			l.calcEstimate(duration)
		} else {
			log.Printf("hold time discarded %s", duration.String())
		}
	}
	l.prevEv = &ev
}
func isTimeSeemsLegit(duration time.Duration) bool {
	return duration.Minutes() >= 15 && duration.Hours() <= 2
}

func (l *HoldTimeEstimateListener) calcEstimate(duration time.Duration) {
	estimate, err := l.estimateRepository.Read()
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
