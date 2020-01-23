package action

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/model"
	"sync"
	"testing"
)

func TestService_Add_DifferentUsers(t *testing.T) {
	service := newInmemService()
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
	service := newInmemService()
	err := service.Pop()
	if err != nil {
		t.Error(err)
	}
	err = service.Add(model.QueueEntity{UserId: "123"})
	if err != nil {
		t.Error(err)
	}
	err = service.Pop()
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
	service := newInmemService()
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
	service := newInmemService()
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
	service := newInmemService()
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

func addUsers(service Service, t *testing.T, start, end int, group *sync.WaitGroup) {
	defer group.Done()

	for i := start; i < end; i++ {
		err := service.Add(model.QueueEntity{UserId: fmt.Sprint(i)})
		if err != nil {
			t.Error(err)
		}
	}
}

func newInmemService() Service {
	return &service{&inmemRepository{model.Queue{}}, sync.Mutex{}}
}

type inmemRepository struct {
	model.Queue
}

func (i *inmemRepository) Save(queue model.Queue) error {
	i.Queue = queue
	return nil
}

func (i *inmemRepository) Read() (model.Queue, error) {
	return i.Queue, nil
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
