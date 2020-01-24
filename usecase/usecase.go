package usecase

import (
	"errors"
	"fmt"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
)

type QueueService interface {
	Add(model.QueueEntity) error
	DeleteById(userId string) error
	Pop() error
	DeleteAll() error
	Show() (model.Queue, error)
}

type service struct {
	queue.Repository
}

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUserErr   = errors.New("no such user")
	QueueIsEmpty    = errors.New("queue is empty")
)

func NewQueueService(repository queue.Repository) QueueService {
	if _, err := repository.Read(); err != nil {
		panic(fmt.Sprintf("can't crete QueueService: %s", err))
	}
	return &service{repository}
}

func (s *service) Pop() error {
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(queue.Entities) == 0 {
		return QueueIsEmpty
	}
	return s.DeleteById(queue.Entities[0].UserId)
}

func (s *service) Add(entity model.QueueEntity) error {
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}

	i := queue.IndexOf(entity.UserId)
	if i != -1 {
		return AlreadyExistErr
	}
	queue.Entities = append(queue.Entities, entity)
	err = s.Repository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteById(userId string) error {
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(queue.Entities) == 0 {
		return QueueIsEmpty
	}
	i := queue.IndexOf(userId)
	if i == -1 {
		return NoSuchUserErr
	}
	queue.Entities = append(queue.Entities[:i], queue.Entities[i+1:]...)
	err = s.Repository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteAll() error {
	q, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(q.Entities) == 0 {
		return QueueIsEmpty
	}
	err = s.Repository.Save(model.Queue{})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Show() (model.Queue, error) {
	return s.Read()
}
