package impl

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/event"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/usecase"
	"log"
	"sync"
	"time"
)

type service struct {
	rep      queue.Repository
	bus      event.QueueChangedEventBus
	mu       sync.Mutex
	slackApi *slack.Client
}

func NewQueueService(repository queue.Repository, queueChangedEventBus event.QueueChangedEventBus) usecase.QueueService {
	if _, err := repository.Read(); err != nil {
		panic(fmt.Sprintf("can't crete QueueService: %s", err))
	}
	return &service{repository, queueChangedEventBus, sync.Mutex{}, nil}
}

func (s *service) Pop(authorUserId string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return "", err
	}
	if len(queue.Entities) == 0 {
		return "", usecase.QueueIsEmpty
	}
	err = s.deleteById(queue.Entities[0].UserId, authorUserId)
	if err != nil {
		return "", err
	}
	return queue.Entities[0].UserId, nil
}

func (s *service) Add(entity model.QueueEntity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return err
	}
	defer func(queueBefore model.Queue) {
		if err == nil {
			go s.emitEvents(entity.UserId, queueBefore, queue)
		}
	}(queue.Copy())

	i := queue.IndexOf(entity.UserId)
	if i != -1 {
		return usecase.AlreadyExistErr
	}
	queue.Entities = append(queue.Entities, entity)
	err = s.rep.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteById(toDelUserId string, authorUserId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deleteById(toDelUserId, authorUserId)

}

//lock must acquired in caller method
func (s *service) deleteById(toDelUserId string, authorUserId string) error {
	queue, err := s.rep.Read()
	if err != nil {
		return err
	}
	defer func(queueBefore model.Queue) {
		if err == nil {
			go s.emitEvents(authorUserId, queueBefore, queue)
		}
	}(queue.Copy())
	if len(queue.Entities) == 0 {
		return usecase.QueueIsEmpty
	}
	i := queue.IndexOf(toDelUserId)
	if i == -1 {
		return usecase.NoSuchUserErr
	}
	queue.Entities = append(queue.Entities[:i], queue.Entities[i+1:]...)
	err = s.rep.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, err := s.rep.Read()
	if err != nil {
		return err
	}
	if len(q.Entities) == 0 {
		return usecase.QueueIsEmpty
	}
	err = s.rep.Save(model.Queue{})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Show() (model.Queue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rep.Read()
}
func (s *service) UpdateOnNewHolder() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, err := s.rep.Read()
	if err != nil {
		return err
	}
	if q.CurHolder() == "" {
		q.HolderIsSleeping = false
		q.HoldTs = time.Time{}
	} else {
		q.HolderIsSleeping = true
		q.HoldTs = time.Now()
	}

	err = s.rep.Save(q)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Pass(holder string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return err
	}
	defer func(queueBefore model.Queue) {
		if err == nil {
			go s.emitEvents(holder, queueBefore, queue)
		}
	}(queue.Copy())
	i := queue.IndexOf(holder)
	if i != 0 {
		return usecase.YouAreNotHolder
	}
	if len(queue.Entities) < 2 {
		return usecase.NoOneToPass
	}
	queue.Entities[0], queue.Entities[1] = queue.Entities[1], queue.Entities[0]
	err = s.rep.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Ack(authorUserId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, err := s.rep.Read()
	if err != nil {
		return err
	}
	if len(q.Entities) == 0 || q.Entities[0].UserId != authorUserId {
		return usecase.YouAreNotHolder
	}
	if !q.HolderIsSleeping {
		return usecase.HolderIsNotSleeping
	}
	q.HolderIsSleeping = false
	err = s.rep.Save(q)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) emitEvents(authorUserId string, before model.Queue, after model.Queue) {
	s.emitNewHolderEvent(before, after, authorUserId)
	s.emitNewSecondEvent(before, after)
}

func (s *service) emitNewSecondEvent(before model.Queue, after model.Queue) {
	secondBefore, secondAfter := "", ""
	if len(before.Entities) > 1 {
		secondBefore = before.Entities[1].UserId
	}
	if len(after.Entities) > 1 {
		secondAfter = after.Entities[1].UserId
	}
	if secondBefore != secondAfter && secondAfter != "" {
		s.bus.Send(model.NewSecondEvent{CurrentSecondUserId: secondAfter})
	}
}
func (s *service) emitNewHolderEvent(before model.Queue, after model.Queue, authorUserId string) {
	holderBefore, holderAfter := "", ""
	if len(before.Entities) > 0 {
		holderBefore = before.Entities[0].UserId
	}
	if len(after.Entities) > 0 {
		holderAfter = after.Entities[0].UserId

		if holderBefore != holderAfter {
			newHolderEvent := model.NewHolderEvent{
				CurrentHolderUserId: holderAfter,
				PrevHolderUserId:    holderBefore,
				AuthorUserId:        authorUserId,
				Ts:                  time.Now(),
			}
			s.bus.Send(newHolderEvent)
			s.notifyNewHolder(newHolderEvent)
		}
	}
}

const waitForAck = time.Minute * 7

func (s *service) notifyNewHolder(newHolderEvent model.NewHolderEvent) {
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

	err = s.sendMsg(curHolder, fmt.Sprintf(i18n.P.MustGetString("your_turn_came"), waitForAck))
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
		s.sendMsgAndLog(holderUserId, "я бы передал твой ход следующему, пока ты спишь, но ты один в очереди")
		return
	}
	if err != nil {
		log.Printf("can't passSleepingHolder %s", err)
		return
	}
	s.sendMsgAndLog(holderUserId, "твой ход передался следующему, пока ты спал")
}

func (s *service) sendMsg(userId, txt string) error {
	if userId == "" {
		log.Printf("sendMsg user id is empty")
		return nil
	}
	_, _, err := s.slackApi.PostMessage(userId,
		slack.MsgOptionText(txt, true),
		slack.MsgOptionAsUser(true),
	)
	return err
}

func (s *service) sendMsgAndLog(userId, txt string) {
	err := s.sendMsg(userId, txt)
	if err != nil {
		log.Printf("can't send %s %s", userId, txt)
	}
}
