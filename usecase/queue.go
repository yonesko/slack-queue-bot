package usecase

import (
	"errors"
	"fmt"
	"github.com/yonesko/slack-queue-bot/event"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"sync"
	"time"
)

type QueueService interface {
	Add(model.QueueEntity) error
	DeleteById(toDelUserId string, authorUserId string) error
	Pop(authorUserId string) (string, error)
	Ack(authorUserId string) error
	Pass(authorUserId string) error
	DeleteAll() error
	Show() (model.Queue, error)
	UpdateNewHolder() error
}

type service struct {
	rep queue.Repository
	bus event.QueueChangedEventBus
	mu  sync.Mutex
}

var (
	AlreadyExistErr     = errors.New("already exist")
	NoSuchUserErr       = errors.New("no such user")
	QueueIsEmpty        = errors.New("queue is empty")
	HolderIsNotSleeping = errors.New("holder is not sleeping")
	YouAreNotHolder     = errors.New("you are not holder")
	NoOneToPass         = errors.New("no one to pass")
)

func NewQueueService(repository queue.Repository, queueChangedEventBus event.QueueChangedEventBus) QueueService {
	if _, err := repository.Read(); err != nil {
		panic(fmt.Sprintf("can't crete QueueService: %s", err))
	}
	return &service{repository, queueChangedEventBus, sync.Mutex{}}
}

func (s *service) Pop(authorUserId string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.rep.Read()
	if err != nil {
		return "", err
	}
	if len(queue.Entities) == 0 {
		return "", QueueIsEmpty
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
		return AlreadyExistErr
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
		return QueueIsEmpty
	}
	i := queue.IndexOf(toDelUserId)
	if i == -1 {
		return NoSuchUserErr
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
		return QueueIsEmpty
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
func (s *service) UpdateNewHolder() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, err := s.rep.Read()
	if err != nil {
		return err
	}
	if q.CurHolder() == "" {
		q.HolderIsSleeping = false
	}
	return nil
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
	i := queue.IndexOf(authorUserId)
	if i != 0 {
		return YouAreNotHolder
	}
	if len(queue.Entities) < 2 {
		return NoOneToPass
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
		return YouAreNotHolder
	}
	if !q.HolderIsSleeping {
		return HolderIsNotSleeping
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
	if secondBefore != secondAfter {
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
			s.bus.Send(model.NewHolderEvent{
				CurrentHolderUserId: holderAfter,
				PrevHolderUserId:    holderBefore,
				AuthorUserId:        authorUserId,
				Ts:                  time.Now(),
			})
		}
	}
}
