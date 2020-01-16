package queue

import "errors"

type Service interface {
	Add(User) error
	Delete(User) error
	DeleteAll() error
	Show() (Queue, error)
}

type service struct {
	Repository
}

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUser      = errors.New("no such user")
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
	i := queue.indexOf(user)
	if i == -1 {
		return NoSuchUser
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
	return service{fileRepository{filename: "slack-queue-bot.db.json"}}
}
