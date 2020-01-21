package queue

import (
	"github.com/yonesko/slack-queue-bot/model"
	"os"
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
	err := repository.Save(model.Queue{Users: []model.User{{Id: "54"}, {Id: "154"}}})
	if err != nil {
		t.Error(err)
	}
	queue, err := repository.Read()
	if err != nil {
		t.Error(err)
	}
	assertState(t, queue, []string{"54", "154"})
	err = repository.Save(model.Queue{Users: []model.User{{Id: "54"}, {Id: "987654"}}})
	if err != nil {
		t.Error(err)
	}
	queue, err = repository.Read()
	if err != nil {
		t.Error(err)
	}
	assertState(t, queue, []string{"54", "987654"})
}

func assertState(t *testing.T, queue model.Queue, userIds []string) {
	if !equals(queue, userIds) {
		t.Errorf("got=%s want=%s", queue, userIds)
	}
}

func equals(queue model.Queue, userIds []string) bool {
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
