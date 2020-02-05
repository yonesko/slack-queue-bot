package usecase

import (
	"errors"
	"fmt"
	"github.com/yonesko/slack-queue-bot/event"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"log"
	"sync"
	"time"
)

type QueueService interface {
	Add(model.QueueEntity) error
	DeleteById(toDelUserId string, authorUserId string) error
	Pop(authorUserId string) (string, error)
	DeleteAll() error
	Show() (model.Queue, error)
}

type service struct {
	queueRepository      queue.Repository
	queueChangedEventBus event.QueueChangedEventBus
	mu                   sync.Mutex
}

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUserErr   = errors.New("no such user")
	QueueIsEmpty    = errors.New("queue is empty")
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
	queue, err := s.queueRepository.Read()
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
	queue, err := s.queueRepository.Read()
	defer func(queueBefore model.Queue) {
		if err == nil {
			s.emitEvent(entity.UserId, queueBefore, queue)
		}
	}(queue.Copy())
	if err != nil {
		return err
	}

	i := queue.IndexOf(entity.UserId)
	if i != -1 {
		return AlreadyExistErr
	}
	queue.Entities = append(queue.Entities, entity)
	err = s.queueRepository.Save(queue)
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
	queue, err := s.queueRepository.Read()
	defer func(queueBefore model.Queue) {
		if err == nil {
			s.emitEvent(authorUserId, queueBefore, queue)
		}
	}(queue.Copy())
	if err != nil {
		return err
	}
	if len(queue.Entities) == 0 {
		return QueueIsEmpty
	}
	i := queue.IndexOf(toDelUserId)
	if i == -1 {
		return NoSuchUserErr
	}
	queue.Entities = append(queue.Entities[:i], queue.Entities[i+1:]...)
	err = s.queueRepository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, err := s.queueRepository.Read()
	if err != nil {
		return err
	}
	if len(q.Entities) == 0 {
		return QueueIsEmpty
	}
	err = s.queueRepository.Save(model.Queue{})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Show() (model.Queue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.queueRepository.Read()
}

//lock must acquired in caller method
func (s *service) updateHoldTs() {
	q, err := s.queueRepository.Read()
	holdTs := time.Now()
	if err != nil {
		log.Printf("updateHoldTs %s", err)
		holdTs = time.Time{}
	}
	q.HoldTs = holdTs
	err = s.queueRepository.Save(q)
	if err != nil {
		log.Printf("updateHoldTs %s", err)
	}
}

func (s *service) emitEvent(authorUserId string, before model.Queue, after model.Queue) {
	holderBefore, holderAfter, secondUserId := "", "", ""
	if len(before.Entities) > 0 {
		holderBefore = before.Entities[0].UserId
	}
	if len(after.Entities) > 0 {
		holderAfter = after.Entities[0].UserId
	}
	if len(after.Entities) > 1 {
		secondUserId = after.Entities[1].UserId
	}
	if holderBefore != holderAfter {
		s.queueChangedEventBus.Send(model.NewHolderEvent{
			CurrentHolderUserId: holderAfter,
			PrevHolderUserId:    holderBefore,
			AuthorUserId:        authorUserId,
			SecondUserId:        secondUserId,
			Ts:                  time.Now(),
		})
		s.updateHoldTs()
	}
}
