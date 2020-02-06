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
		log.Printf("can't UpdateOnNewHolder, return %s", err)
		return
	}
	if curHolder == "" {
		log.Printf("holder is empty, return")
		return
	}

	go func() {
		txt := fmt.Sprintf(i18n.L.MustGet("your_turn_came"), waitForAck)
		err := s.gateway.Send(curHolder, txt)
		if err != nil {
			log.Printf("can't send %s '%s' %s", curHolder, txt, err)
			return
		}
		time.AfterFunc(waitForAck, func() { s.passFromSleepingHolder(curHolder) })
	}()
}

func (s *service) passFromSleepingHolder(holderUserId string) {
	err := s.PassFromSleepingHolder(holderUserId)
	if err == usecase.HolderIsNotSleeping {
		log.Printf("passFromSleepingHolder %s", err)
		return
	}
	if err == usecase.YouAreNotHolder {
		log.Printf("passFromSleepingHolder %s", err)
		return
	}
	if err == usecase.NoOneToPass {
		s.gateway.SendAndLog(holderUserId, "Я бы передал твой ход следующему, пока ты спишь, но ты один в очереди")
		return
	}
	if err != nil {
		log.Printf("can't passFromSleepingHolder %s", err)
		return
	}
	s.gateway.SendAndLog(holderUserId, "Твой ход передался следующему, пока ты спал")
}
