package queue

import (
	"fmt"
	"sync"
	"testing"
)

func TestService_Add_DifferentUsers(t *testing.T) {
	service := newInmemService()
	err := service.Add(User{Id: "123"})
	if err != nil {
		t.Error(err)
	}
	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	equals(queue, []string{"123"})
	_ = service.Add(User{Id: "ABC"})
	_ = service.Add(User{Id: "ABCD"})
	equals(queue, []string{"123", "ABC", "ABCD"})
}

func TestService_Pop(t *testing.T) {
	service := newInmemService()
	err := service.Pop()
	if err != nil {
		t.Error(err)
	}
	err = service.Add(User{Id: "123"})
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
	err = service.Add(User{Id: "123"})
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
	err := service.Add(User{Id: "123"})
	if err != nil {
		t.Error(err)
	}
	err = service.Add(User{Id: "123"})
	if err == nil || err.Error() != "already exist" {
		t.Error("must be already exist")
	}
}

func TestFileRepositoryParallel(t *testing.T) {
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
	if len(queue.Users) != chunks*workers {
		t.Errorf("must be %d, got %d", chunks*workers, len(queue.Users))
	}
}

func addUsers(service Service, t *testing.T, start, end int, group *sync.WaitGroup) {
	defer group.Done()

	for i := start; i < end; i++ {
		err := service.Add(User{Id: fmt.Sprint(i)})
		if err != nil {
			t.Error(err)
		}
	}
}

func newInmemService() Service {
	return &service{&inmemRepository{Queue{}}, sync.Mutex{}}
}

type inmemRepository struct {
	Queue
}

func (i *inmemRepository) Save(queue Queue) error {
	i.Queue = queue
	return nil
}

func (i *inmemRepository) Read() (Queue, error) {
	return i.Queue, nil
}
