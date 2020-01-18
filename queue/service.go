package queue

import (
	"errors"
	"os"
)

type Service interface {
	Add(User) error
	Delete(User) error
	Pop() error
	DeleteAll() error
	Show() (Queue, error)
}

type service struct {
	Repository
}

func (s service) Pop() error {
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

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUserErr   = errors.New("no such user")
)

func (s service) Add(user User) error {
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}

	i := queue.indexOf(user)
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

func (s service) Delete(user User) error {
	queue, err := s.Repository.Read()
	if err != nil {
		return err
	}
	if len(queue.Users) == 0 {
		return nil
	}
	i := queue.indexOf(user)
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

func (s service) DeleteAll() error {
	err := s.Repository.Save(Queue{})
	if err != nil {
		return err
	}
	return nil
}

func (s service) Show() (Queue, error) {
	return s.Read()
}

func NewService() Service {
	err := os.Mkdir("db", os.ModePerm)
	if err != nil {
		panic(err)
	}
	return service{fileRepository{filename: "db/slack-queue-bot.db.json"}}
}
