package usecase

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/stretchr/testify/assert"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/model"
	queuemock "github.com/yonesko/slack-queue-bot/queue/mock"
	"sync"
	"testing"
	"time"
)

func TestService_Add_DifferentUsers(t *testing.T) {
	service := mockService()
	err := service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	queue, err := service.Show()
	assert.Nil(t, err)
	equals(queue, []string{"123"})
	_ = service.Add(model.QueueEntity{UserId: "ABC"})
	_ = service.Add(model.QueueEntity{UserId: "ABCD"})
	equals(queue, []string{"123", "ABC", "ABCD"})
}

//noinspection GoUnhandledErrorResult
func TestService_HoldTs(t *testing.T) {
	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	defer patch.Unpatch()
	service := mockService()
	service.Add(model.QueueEntity{UserId: "123"})
	queue, _ := service.Show()
	assert.Equal(t, now, queue.HoldTs)
	service.Add(model.QueueEntity{UserId: "2"})
	service.Add(model.QueueEntity{UserId: "3"})
	assert.Equal(t, now, queue.HoldTs)
	service.DeleteById("2", "2")
	assert.Equal(t, now, queue.HoldTs)
	now = time.Now().Add(time.Hour)
	service.DeleteById("123", "123")
	queue, _ = service.Show()
	assert.Equal(t, now.String(), queue.HoldTs.String())
}

func TestService_Pop(t *testing.T) {
	service := mockService()
	_, err := service.Pop("123")
	assert.Equal(t, QueueIsEmpty, err)
	err = service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	deletedUserId, err := service.Pop("123")
	assert.Nil(t, err)
	assert.Equal(t, "123", deletedUserId, "wrong deletedUserId: %s", deletedUserId)
	queue, err := service.Show()
	assert.Nil(t, err)
	equals(queue, []string{})
}

func TestService_DeleteAll(t *testing.T) {
	service := mockService()
	err := service.DeleteAll()
	if err != QueueIsEmpty {
		t.Error(err)
	}
	err = service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	queue, err := service.Show()
	assert.Nil(t, err)
	equals(queue, []string{"123"})
}

func TestService_Add_Idempotent(t *testing.T) {
	service := mockService()
	err := service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	err = service.Add(model.QueueEntity{UserId: "123"})
	if err == nil || err.Error() != "already exist" {
		t.Error("must be already exist")
	}
}

func TestNoRaceConditionsInService(t *testing.T) {
	service := mockService()
	group := &sync.WaitGroup{}
	chunks, workers := 100, 100
	for i := 0; i < workers; i++ {
		group.Add(1)
		go addUsers(service, t, i*chunks, (i+1)*chunks, group)
	}
	group.Wait()

	queue, err := service.Show()
	assert.Nil(t, err)
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

func mockService() *service {
	return &service{
		queuemock.NewQueueRepositoryMock(),
		&eventmock.QueueChangedEventBus{Inbox: []interface{}{}},
		sync.Mutex{},
	}
}
