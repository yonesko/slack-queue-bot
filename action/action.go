package action

import (
	"errors"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
)

var (
	AlreadyExistErr = errors.New("already exist")
	NoSuchUserErr   = errors.New("no such user")
)

type AddToQueue interface {
	Do(entity model.QueueEntity) error
}

type addToQueue struct {
	queueRepository queue.Repository
}

func (a *addToQueue) Do(entity model.QueueEntity) error {
	queue, err := a.queueRepository.Read()
	if err != nil {
		return err
	}

	i := queue.IndexOf(entity)
	if i != -1 {
		return AlreadyExistErr
	}
	queue.Entities = append(queue.Entities, entity)
	err = a.queueRepository.Save(queue)
	if err != nil {
		return err
	}
	return nil
}
