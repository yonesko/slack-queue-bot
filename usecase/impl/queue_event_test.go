package impl

import (
	"github.com/stretchr/testify/assert"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	queuemock "github.com/yonesko/slack-queue-bot/queue/mock"
	"github.com/yonesko/slack-queue-bot/usecase"
	"sync"
	"testing"
	"time"
)

func TestNewHolderEventAddToEmptyQueue(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{})
	err := service.Add(model.QueueEntity{UserId: "123"})
	time.Sleep(time.Millisecond)
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).CurrentHolderUserId)
	assert.Equal(t, "", bus.Inbox[0].(model.NewHolderEvent).PrevHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).AuthorUserId)
}

//noinspection GoUnhandledErrorResult
func Test_NewHolderEvent_TheSecond_HolderRemoved(t *testing.T) {
	bus, service := buildQueueServiceAndBus(model.Queue{})
	service.Add(model.QueueEntity{UserId: "123"})
	service.Add(model.QueueEntity{UserId: "abc"})
	service.Add(model.QueueEntity{UserId: "z"})

	err := service.DeleteById("123", "123")
	assert.Nil(t, err)
	time.Sleep(time.Millisecond)
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "z"})
}

//noinspection GoUnhandledErrorResult
func Test_NewHolderEvent_TheSecond_SecondRemoved(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}, {"z"}}})

	err := service.DeleteById("abc", "123")
	time.Sleep(time.Millisecond * 10)
	assert.Nil(t, err)
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "z"})
}

//noinspection GoUnhandledErrorResult
func TestNewHolderEventSelfDeleteHolder(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}})

	err := service.DeleteById("123", "123")
	time.Sleep(time.Millisecond * 10)
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "abc", bus.Inbox[0].(model.NewHolderEvent).CurrentHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).PrevHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).AuthorUserId)
}

//noinspection GoUnhandledErrorResult
func TestNewHolderEventForceDeleteHolder(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}, {"z"}}})

	assert.Nil(t, service.DeleteById("123", "jhgfdvxc"))
	time.Sleep(time.Millisecond * 10)
	assert.True(t, containsNewHolderEvent(bus.Inbox, "abc", "jhgfdvxc", "123"))
}

func TestNewHolderEventSelfDeleteNotHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("abc", "abc")
	assert.Nil(t, err)
	assert.Empty(t, bus.Inbox)
}

//noinspection GoUnhandledErrorResult
func Test_Pass_emits_events(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"123"}}})
	service.Pass("123")
	assert.Empty(t, bus.Inbox)
	//
	service.Add(model.QueueEntity{UserId: "a"})
	service.Pass("123")
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "123"})
	time.Sleep(time.Millisecond * 5)
	assert.True(t, containsNewHolderEvent(bus.Inbox, "a", "123", "123"))
	//
	bus.Inbox = nil
	service.Add(model.QueueEntity{UserId: "b"})
	service.Add(model.QueueEntity{UserId: "c"})
	time.Sleep(time.Millisecond * 5)
	assert.Empty(t, bus.Inbox)
}

func Test_delete_all_emits_DeletedEvent(t *testing.T) {
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"1"}, {"2"}, {"3"}}})
	assert.Nil(t, service.DeleteAll("5"))
	assert.Contains(t, bus.Inbox, model.DeletedEvent{"5", "1"})
	assert.Contains(t, bus.Inbox, model.DeletedEvent{"5", "2"})
	assert.Contains(t, bus.Inbox, model.DeletedEvent{"5", "3"})
}

func Test_self_delete_all_emits_DeletedEvent(t *testing.T) {
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"1"}, {"2"}, {"3"}}})
	assert.Nil(t, service.DeleteAll("2"))
	assert.Contains(t, bus.Inbox, model.DeletedEvent{"2", "1"})
	assert.Contains(t, bus.Inbox, model.DeletedEvent{"2", "3"})
}

func Test_pop_emits_DeletedEvent(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"1"}, {"2"}, {"3"}}})
	_, err := service.Pop("2")
	assert.Nil(t, err)
	assert.Contains(t, bus.Inbox, model.DeletedEvent{"2", "1"})
}

func Test_no_new_holder_event_when_delete_not_holder(t *testing.T) {
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}})

	assert.Nil(t, service.DeleteById("abc", "jjfftg"))
	assert.Condition(t, func() bool {
		for _, e := range bus.Inbox {
			if _, ok := e.(model.NewHolderEvent); ok {
				return false
			}
		}
		return true
	})
}

func TestNewHolderEventPopAnotherUser(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}})

	_, err := service.Pop("abc")
	time.Sleep(time.Millisecond * 10)
	assert.Nil(t, err)
	containsNewHolderEvent(bus.Inbox, "abc", "abc", "123")
}

func TestNewHolderEventPopOnEmpty(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	_, err := service.Pop("123")
	assert.Equal(t, usecase.QueueIsEmpty, err)
	assert.Empty(t, bus.Inbox)
}

func Test_NewHolderEvent_CheckForEvents_on_PassFromSleepingHolder(t *testing.T) {
	i18n.TestInit()
	bus, service := buildQueueServiceAndBus(model.Queue{})

	assert.Equal(t, usecase.HolderIsNotSleeping, service.PassFromSleepingHolder("5653"))
	assert.Empty(t, bus.Inbox)
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "4"}))
	assert.Equal(t, usecase.NoOneToPass, service.PassFromSleepingHolder("4"))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "6"}))
	assert.Nil(t, service.PassFromSleepingHolder("4"))
	//6 4
	time.Sleep(time.Millisecond * 5)
	containsNewHolderEvent(bus.Inbox, "6", "4", "4")
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "4"})
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "1"}))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "17"}))
	//6 4 1 17
	assert.Equal(t, usecase.YouAreNotHolder, service.PassFromSleepingHolder("4"))
	assert.Nil(t, service.PassFromSleepingHolder("6"))
	//4 6 1 17
	time.Sleep(time.Millisecond * 5)
	containsNewHolderEvent(bus.Inbox, "4", "6", "6")
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "6"})
}

func buildQueueServiceAndBus(queue model.Queue) (*eventmock.QueueChangedEventBus, *service) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{queue}
	service := &service{&queueRepository, &bus, sync.Mutex{}, gateway.Mock{}}
	return &bus, service
}

func containsNewHolderEvent(inbox []interface{}, cur, au, prev string) bool {
	for _, e := range inbox {
		if event, ok := e.(model.NewHolderEvent); ok {
			if event.CurrentHolderUserId == cur && event.AuthorUserId == au && event.PrevHolderUserId == prev {
				return true
			}
		}
	}
	return false
}
