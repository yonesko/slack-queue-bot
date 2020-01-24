package usecase

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue/mock"
	"sync"
	"testing"
)

func TestService_Add_DifferentUsers(t *testing.T) {
	service := &service{mock.NewQueueRepositoryMock()}
	err := service.Add(model.QueueEntity{UserId: "123"})
	if err != nil {
		t.Error(err)
	}
	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	equals(queue, []string{"123"})
	_ = service.Add(model.QueueEntity{UserId: "ABC"})
	_ = service.Add(model.QueueEntity{UserId: "ABCD"})
	equals(queue, []string{"123", "ABC", "ABCD"})
}

func TestService_Pop(t *testing.T) {
	service := &service{mock.NewQueueRepositoryMock()}
	_, err := service.Pop()
	if err != nil {
		t.Error(err)
	}
	err = service.Add(model.QueueEntity{UserId: "123"})
	if err != nil {
		t.Error(err)
	}
	_, err = service.Pop()
	if err != nil {
		t.Error(err)
	}
	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	equals(queue, []string{})
}

func TestService_DeleteAll(t *testing.T) {
	service := &service{mock.NewQueueRepositoryMock()}
	err := service.DeleteAll()
	if err != nil {
		t.Error(err)
	}
	err = service.Add(model.QueueEntity{UserId: "123"})
	if err != nil {
		t.Error(err)
	}
	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	equals(queue, []string{"123"})
}

func TestService_Add_Idempotent(t *testing.T) {
	service := &service{mock.NewQueueRepositoryMock()}
	err := service.Add(model.QueueEntity{UserId: "123"})
	if err != nil {
		t.Error(err)
	}
	err = service.Add(model.QueueEntity{UserId: "123"})
	if err == nil || err.Error() != "already exist" {
		t.Error("must be already exist")
	}
}

func TestNoRaceConditionsInService(t *testing.T) {
	service := &service{mock.NewQueueRepositoryMock()}
	group := &sync.WaitGroup{}
	chunks, workers := 100, 100
	for i := 0; i < workers; i++ {
		group.Add(1)
		go addUsers(service, t, i*chunks, (i+1)*chunks, group)
	}
	group.Wait()

	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	if len(queue.Entities) != chunks*workers {
		t.Errorf("must be %d, got %d", chunks*workers, len(queue.Entities))
	}
}

func addUsers(service QueueService, t *testing.T, start, end int, group *sync.WaitGroup) {
	defer group.Done()

	for i := start; i < end; i++ {
		err := service.Add(model.QueueEntity{UserId: fmt.Sprint(i)})
		if err != nil {
			t.Error(err)
		}
	}
}

func equals(queue model.Queue, userIds []string) bool {
	if len(queue.Entities) != len(userIds) {
		return false
	}

	for i := range userIds {
		if userIds[i] != queue.Entities[i].UserId {
			return false
		}
	}

	return true
}
