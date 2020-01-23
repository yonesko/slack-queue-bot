package mock

import "github.com/yonesko/slack-queue-bot/model"

type QueueRepositoryMock struct {
	model.Queue
}

func (i *QueueRepositoryMock) Save(queue model.Queue) error {
	i.Queue = queue
	return nil
}

func (i *QueueRepositoryMock) Read() (model.Queue, error) {
	return i.Queue, nil
}
