package mock

import (
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
)

type QueueRepository struct {
	model.Queue
}

func NewQueueRepositoryMock() queue.Repository {
	return &QueueRepository{model.Queue{}}
}

func (i *QueueRepository) Save(queue model.Queue) error {
	i.Queue = queue
	return nil
}

func (i *QueueRepository) Read() (model.Queue, error) {
	return i.Queue, nil
}
