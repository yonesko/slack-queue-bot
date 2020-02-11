package impl

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/stretchr/testify/assert"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue/mock"
	"github.com/yonesko/slack-queue-bot/usecase"
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
	i18n.TestInit()
	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	defer patch.Unpatch()
	service := mockService()
	service.Add(model.QueueEntity{UserId: "123"})
	time.Sleep(time.Millisecond * 5)
	queue, _ := service.Show()
	assert.Equal(t, now, queue.HoldTs)
	service.Add(model.QueueEntity{UserId: "2"})
	service.Add(model.QueueEntity{UserId: "3"})
	assert.Equal(t, now, queue.HoldTs)
	service.DeleteById("2", "2")
	assert.Equal(t, now, queue.HoldTs)
	now = time.Now().Add(time.Hour)
	service.DeleteById("123", "123")
	time.Sleep(time.Millisecond * 5)
	queue, _ = service.Show()
	assert.Equal(t, now.String(), queue.HoldTs.String())
}

func TestService_Pop(t *testing.T) {
	service := mockService()
	_, err := service.Pop("123")
	assert.Equal(t, usecase.QueueIsEmpty, err)
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
	err := service.DeleteAll("")
	if err != usecase.QueueIsEmpty {
		t.Error(err)
	}
	err = service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	queue, err := service.Show()
	assert.Nil(t, err)
	equals(queue, []string{"123"})
}

func Test_Pass(t *testing.T) {
	i18n.TestInit()
	service := mockService()
	assert.Equal(t, usecase.QueueIsEmpty, service.Pass(""))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "123"}))
	assert.Equal(t, usecase.NoSuchUserErr, service.Pass("333"))
	assert.Equal(t, usecase.NoOneToPass, service.Pass("123"))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "456"}))
	assert.Nil(t, service.Pass("123"))
	queue, _ := service.Show()
	equals(queue, []string{"456", "123"})
	//
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "a"}))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "b"}))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "c"}))
	equals(queue, []string{"456", "123", "a", "b", "c"})
	assert.Nil(t, service.Pass("123"))
	queue, _ = service.Show()
	equals(queue, []string{"456", "a", "123", "b", "c"})
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
	i18n.TestInit()
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

//noinspection GoUnhandledErrorResult
func TestAck(t *testing.T) {
	service := mockService()
	service.Add(model.QueueEntity{UserId: "1"})
	time.Sleep(time.Millisecond * 5)
	queue, _ := service.Show()
	assert.True(t, queue.HolderIsSleeping)
	assert.Equal(t, usecase.YouAreNotHolder, service.Ack("5"))
	queue, _ = service.Show()
	assert.True(t, queue.HolderIsSleeping)
	assert.Nil(t, service.Ack("1"))
	assert.Equal(t, usecase.HolderIsNotSleeping, service.Ack("1"))
	queue, _ = service.Show()
	assert.False(t, queue.HolderIsSleeping)
	service.Add(model.QueueEntity{UserId: "6"})
	queue, _ = service.Show()
	assert.False(t, queue.HolderIsSleeping)
	service.Add(model.QueueEntity{UserId: "6"})
}

func TestService_UpdateNewHolder(t *testing.T) {
	now := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return now })
	defer patch.Unpatch()

	service := mockService()
	assert.Nil(t, service.UpdateOnNewHolder())
	queue, _ := service.Show()
	assert.False(t, queue.HolderIsSleeping)
	assert.Zero(t, queue.HoldTs)
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "1"}))
	assert.Nil(t, service.UpdateOnNewHolder())
	queue, _ = service.Show()
	assert.True(t, queue.HolderIsSleeping)
	assert.Equal(t, now, queue.HoldTs)
}

func TestService_PassFromSleepingHolder(t *testing.T) {
	i18n.TestInit()
	service := mockService()
	assert.Equal(t, usecase.HolderIsNotSleeping, service.PassFromSleepingHolder("5653"))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "4"}))
	assert.Equal(t, usecase.NoOneToPass, service.PassFromSleepingHolder("4"))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "6"}))
	assert.Nil(t, service.PassFromSleepingHolder("4"))
	queue, _ := service.Show()
	equals(queue, []string{"6", "4"})
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "1"}))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "17"}))
	equals(queue, []string{"6", "4", "1", "17"})
	assert.Equal(t, usecase.YouAreNotHolder, service.PassFromSleepingHolder("4"))
	assert.Nil(t, service.PassFromSleepingHolder("6"))
	equals(queue, []string{"4", "6", "1", "17"})
}

func addUsers(service usecase.QueueService, t *testing.T, start, end int, group *sync.WaitGroup) {
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
		&mock.QueueRepository{model.Queue{}},
		&eventmock.QueueChangedEventBus{Inbox: []interface{}{}},
		sync.Mutex{},
		gateway.Mock{},
	}
}
