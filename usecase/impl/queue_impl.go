package impl

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/event"
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/usecase"
	"sync"
	"time"
)

type service struct {
	rep     queue.Repository
	bus     event.QueueChangedEventBus
	mu      sync.Mutex
	gateway gateway.Gateway
}

func NewQueueService(repository queue.Repository, queueChangedEventBus event.QueueChangedEventBus, gateway gateway.Gateway) usecase.QueueService {
	if _, err := repository.Read(); err != nil {
		panic(fmt.Sprintf("can't crete QueueService: %s", err))
	}
	return &service{repository, queueChangedEventBus, sync.Mutex{}, gateway}
}

func (s *service) Pass(authorUserId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return err
	}
	defer func(queueBefore model.Queue) {
		if err == nil {
			s.emitEvents(authorUserId, queueBefore, queue)
		}
	}(queue.Copy())
	if len(queue.Entities) == 0 {
		return usecase.QueueIsEmpty
	}
	i := queue.IndexOf(authorUserId)
	if i == -1 {
		return usecase.NoSuchUserErr
	}
	nexti := i + 1
	if nexti >= len(queue.Entities) {
		return usecase.NoOneToPass
	}
	queue.Entities[i], queue.Entities[nexti] = queue.Entities[nexti], queue.Entities[i]
	err = s.rep.Save(queue)
	if err != nil {
		return err
	}
	return nil
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
			s.emitEvents(entity.UserId, queueBefore, queue)
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
			s.emitEvents(authorUserId, queueBefore, queue)
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

func (s *service) DeleteAll(authorUserId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return err
	}
	defer func(queueBefore model.Queue) {
		if err == nil {
			s.emitEvents(authorUserId, queueBefore, queue)
		}
	}(queue.Copy())
	if len(queue.Entities) == 0 {
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

//lock must acquired in caller method
func (s *service) UpdateOnNewHolder() error {
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

func (s *service) PassFromSleepingHolder(holder string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return err
	}
	defer func(queueBefore model.Queue) {
		if err == nil {
			s.emitEvents(holder, queueBefore, queue)
		}
	}(queue.Copy())
	if !queue.HolderIsSleeping {
		return usecase.HolderIsNotSleeping
	}
	if queue.CurHolder() != holder {
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
			go s.bus.Send(newHolderEvent)
			s.notifyNewHolderAndWaitForAck(newHolderEvent)
		}
	}
}
