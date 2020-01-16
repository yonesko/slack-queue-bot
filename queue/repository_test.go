package queue

import "testing"

func TestFileRepository(t *testing.T) {
	repository := fileRepository{"slack-queue-bot.test-db.json"}
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

func assertState(t *testing.T, queue Queue, userIds []string) {
	if !equals(queue, userIds) {
		t.Errorf("got=%s want=%s", queue, userIds)
	}
}

func newInmemService() Service {
	return service{&inmemRepository{Queue{}}}
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
