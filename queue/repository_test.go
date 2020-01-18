package queue

import (
	"fmt"
	"os"
	"strconv"
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
	group.Add(1)

	addUsers(service, t, 0, 1000, group)
	group.Wait()

	queue, err := service.Show()
	if err != nil {
		t.Error(err)
	}
	if len(queue.Users) != 1000 {
		t.Error("must be all")
	}
	for i := 0; i < 1000; i++ {
		queue.Users[i].Id = strconv.Itoa(i)
	}
}

func addUsers(service Service, t *testing.T, start, end int, group *sync.WaitGroup) {
	for i := start; i < end; i++ {
		err := service.Add(User{Id: fmt.Sprint(i)})
		if err != nil {
			t.Error(err)
		}
	}
	group.Done()
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
