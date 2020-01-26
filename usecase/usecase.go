package usecase

import (
	"errors"
	"fmt"
	"github.com/yonesko/slack-queue-bot/event"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
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
	return &service{repository, queueChangedEventBus}
}

func (s *service) Pop(authorUserId string) (string, error) {
	queue, err := s.queueRepository.Read()
	if err != nil {
		return "", err
	}
	if len(queue.Entities) == 0 {
		return "", QueueIsEmpty
	}
	err = s.DeleteById(queue.Entities[0].UserId, authorUserId)
	if err != nil {
		return "", err
	}
	return queue.Entities[0].UserId, nil
}

func (s *service) Add(entity model.QueueEntity) error {
	queue, err := s.queueRepository.Read()
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
	queue, err := s.queueRepository.Read()
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
	prevHolderUserId := ""
	if i == 0 {
		prevHolderUserId = toDelUserId
	}
	queue.Entities = append(queue.Entities[:i], queue.Entities[i+1:]...)
	err = s.queueRepository.Save(queue)
	if err != nil {
		return err
	}
	if i == 0 {
		curr := ""
		if len(queue.Entities) > 0 {
			curr = queue.Entities[0].UserId
		}
		go s.queueChangedEventBus.Send(event.NewHolderEvent{
			CurrentHolderUserId: curr,
			PrevHolderUserId:    prevHolderUserId,
			AuthorUserId:        authorUserId,
		})
	}
	return nil
}

func (s *service) DeleteAll() error {
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
	return s.queueRepository.Read()
}
