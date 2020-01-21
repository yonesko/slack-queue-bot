package queue

import (
	"errors"
	"fmt"
	"github.com/yonesko/slack-queue-bot/model"
	"sync"
)

type Service interface {
	Add(model.User) error
	Delete(model.User) error
	Pop() error
	DeleteAll() error
	Show() (model.Queue, error)
}

type service struct {
	Repository
	mu sync.Mutex
}

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUserErr   = errors.New("no such user")
)

func NewService() Service {
	repository := newFileRepository()
	if _, err := repository.Read(); err != nil {
		panic(fmt.Sprintf("can't crete Service: %s", err))
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
	if len(queue.Users) == 0 {
		return nil
	}
	queue.Users = queue.Users[1:]
	err = s.Repository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Add(user model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}

	i := queue.IndexOf(user)
	if i != -1 {
		return AlreadyExistErr
	}
	queue.Users = append(queue.Users, user)
	err = s.Repository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Delete(user model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(queue.Users) == 0 {
		return nil
	}
	i := queue.IndexOf(user)
	if i == -1 {
		return NoSuchUserErr
	}
	queue.Users = append(queue.Users[:i], queue.Users[i+1:]...)
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
