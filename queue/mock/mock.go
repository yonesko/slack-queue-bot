package mock

import (
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
)

type queueRepositoryMock struct {
	model.Queue
}

func NewQueueRepositoryMock() queue.Repository {
	return &queueRepositoryMock{model.Queue{}}
}

func (i *queueRepositoryMock) Save(queue model.Queue) error {
	i.Queue = queue
	return nil
}

func (i *queueRepositoryMock) Read() (model.Queue, error) {
	return i.Queue, nil
}
