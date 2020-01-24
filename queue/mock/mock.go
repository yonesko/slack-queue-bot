package mock

import (
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
)

type queueRepository struct {
	model.Queue
}

func NewQueueRepositoryMock() queue.Repository {
	return &queueRepository{model.Queue{}}
}

func (i *queueRepository) Save(queue model.Queue) error {
	i.Queue = queue
	return nil
}

func (i *queueRepository) Read() (model.Queue, error) {
	return i.Queue, nil
}
