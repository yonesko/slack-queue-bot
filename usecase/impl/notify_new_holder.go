package impl

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/usecase"
	"log"
	"time"
)

const waitForAck = time.Minute * 7

func (s *service) notifyNewHolderAndWaitForAck(newHolderEvent model.NewHolderEvent) {
	curHolder := newHolderEvent.CurrentHolderUserId
	err := s.UpdateOnNewHolder()
	if err != nil {
		log.Printf("can't UpdateOnNewHolder, return")
		return
	}
	if curHolder == "" {
		log.Printf("holder is empty, return")
		return
	}

	err = s.gateway.Send(curHolder, fmt.Sprintf(i18n.P.MustGet("your_turn_came"), waitForAck))
	if err != nil {
		log.Printf("can't notify holder %s: %s", curHolder, err)
		return
	}

	time.AfterFunc(waitForAck, func() { s.passSleepingHolder(curHolder) })
}

func (s *service) passSleepingHolder(holderUserId string) {
	err := s.Pass(holderUserId)
	if err == usecase.YouAreNotHolder {
		log.Printf("passSleepingHolder %s", err)
		return
	}
	if err == usecase.NoOneToPass {
		s.gateway.SendAndLog(holderUserId, "я бы передал твой ход следующему, пока ты спишь, но ты один в очереди")
		return
	}
	if err != nil {
		log.Printf("can't passSleepingHolder %s", err)
		return
	}
	s.gateway.SendAndLog(holderUserId, "твой ход передался следующему, пока ты спал")
}
