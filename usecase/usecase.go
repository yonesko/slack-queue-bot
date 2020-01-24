package usecase

import (
	"errors"
	"fmt"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"sync"
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
	mu sync.Mutex
}

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUserErr   = errors.New("no such user")
)

func NewQueueService(repository queue.Repository) QueueService {
	if _, err := repository.Read(); err != nil {
		panic(fmt.Sprintf("can't crete QueueService: %s", err))
	}
	return &service{repository, sync.Mutex{}}
}

func (s *service) Pop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(queue.Entities) == 0 {
		return nil
	}
	queue.Entities = queue.Entities[1:]
	err = s.Repository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Add(entity model.QueueEntity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(queue.Entities) == 0 {
		return nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.Repository.Save(model.Queue{})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Show() (model.Queue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Read()
}
