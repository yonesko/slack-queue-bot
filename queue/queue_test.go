package queue

import "testing"

type inmemRepository struct {
	Queue
}

func (i *inmemRepository) Save(queue Queue) {
	i.Queue = queue
}

func (i *inmemRepository) Read() Queue {
	return i.Queue
}

func TestAdd(t *testing.T) {
	service := newInmemService()
	if len(service.Show().users) != 0 {
		t.Error("must be 0")
	}
	service.Add(User{Id: "123"})
	assertState(t, service.Show(), []string{"123"})
	//no 2 times
	service.Add(User{Id: "123"})
	service.Add(User{Id: "123"})
	assertState(t, service.Show(), []string{"123"})
	service.Add(User{Id: "ABC"})
	assertState(t, service.Show(), []string{"123", "ABC"})
	//no 2 times again
	service.Add(User{Id: "123"})
	assertState(t, service.Show(), []string{"123", "ABC"})
}

func TestDelete(t *testing.T) {
	service := newInmemService()
	//idempotent
	service.Delete(User{Id: "123"})
	service.Delete(User{Id: "123"})
	//
	service.Add(User{Id: "123"})
	service.Delete(User{Id: "123"})
	assertState(t, service.Show(), []string{})
	//
	service.Add(User{Id: "123"})
	service.Add(User{Id: "ABC"})
	assertState(t, service.Show(), []string{"123", "ABC"})
	service.Delete(User{Id: "123"})
	assertState(t, service.Show(), []string{"ABC"})
	service.Delete(User{Id: "ABC"})
	assertState(t, service.Show(), []string{})
}
func assertState(t *testing.T, queue Queue, userIds []string) {
	if !equals(queue, userIds) {
		t.Errorf("got=%s want=%s", queue, userIds)
	}
}

func newInmemService() Service {
	return service{&inmemRepository{Queue{}}}
}

func equals(queue Queue, userIds []string) bool {
	if len(queue.users) != len(userIds) {
		return false
	}

	for i := range userIds {
		if userIds[i] != queue.users[i].Id {
			return false
		}
	}

	return true

}
