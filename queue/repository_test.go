package queue

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

func init() {
	err := os.RemoveAll("db")
	if err != nil {
		panic(err)
	}
}

func TestFileRepository(t *testing.T) {
	repository := newFileRepository()
	err := repository.Save(Queue{Users: []User{{Id: "54"}, {Id: "154"}}})
	if err != nil {
		t.Error(err)
	}
	queue, err := repository.Read()
	if err != nil {
		t.Error(err)
	}
	assertState(t, queue, []string{"54", "154"})
	err = repository.Save(Queue{Users: []User{{Id: "54"}, {Id: "987654"}}})
	if err != nil {
		t.Error(err)
	}
	queue, err = repository.Read()
	if err != nil {
		t.Error(err)
	}
	assertState(t, queue, []string{"54", "987654"})
}

func TestFileRepositoryParallel(t *testing.T) {
	service := NewService()
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
		t.Errorf("must be all: %d", len(queue.Users))
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

func assertState(t *testing.T, queue Queue, userIds []string) {
	if !equals(queue, userIds) {
		t.Errorf("got=%s want=%s", queue, userIds)
	}
}

func equals(queue Queue, userIds []string) bool {
	if len(queue.Users) != len(userIds) {
		return false
	}

	for i := range userIds {
		if userIds[i] != queue.Users[i].Id {
			return false
		}
	}

	return true

}
