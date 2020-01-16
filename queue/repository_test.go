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
	assertState(t, service.Show(), []string{})
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

func TestAdd_Many(t *testing.T) {
	service := newInmemService()
	assertState(t, service.Show(), []string{})
	N := 1000
	for i := 0; i < N; i++ {
		service.Add(User{Id: string(i)})
	}
	if len(service.Show().Users) != N {
		t.Error()
	}
	N2 := N - 647
	for i := 0; i < N2; i++ {
		service.Delete(User{Id: string(i)})
	}
	if len(service.Show().Users) != N-N2 {
		t.Error()
	}
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

func TestFileRepository(t *testing.T) {
	repository := fileRepository{"slack-queue-bot.test-db.json"}
	repository.Save(Queue{Users: []User{{Id: "54"}, {Id: "154"}}})
	assertState(t, repository.Read(), []string{"54", "154"})
	repository.Save(Queue{Users: []User{{Id: "54"}, {Id: "987654"}}})
	assertState(t, repository.Read(), []string{"54", "987654"})
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
