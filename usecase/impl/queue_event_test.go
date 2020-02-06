package impl

import (
	"github.com/stretchr/testify/assert"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/model"
	queuemock "github.com/yonesko/slack-queue-bot/queue/mock"
	"github.com/yonesko/slack-queue-bot/usecase"
	"sync"
	"testing"
)

func TestNewHolderEventAddToEmptyQueue(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).CurrentHolderUserId)
	assert.Equal(t, "", bus.Inbox[0].(model.NewHolderEvent).PrevHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).AuthorUserId)
}

func Test_NewHolderEvent_TheSecond_HolderRemoved(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}, {"z"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("123", "123")
	assert.Nil(t, err)
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "z"})
}
func Test_NewHolderEvent_TheSecond_SecondRemoved(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}, {"z"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("abc", "123")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "z", bus.Inbox[0].(model.NewSecondEvent).CurrentSecondUserId)
}
func TestNewHolderEventSelfDeleteHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("123", "123")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "abc", bus.Inbox[0].(model.NewHolderEvent).CurrentHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).PrevHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).AuthorUserId)
}

func TestNewHolderEventForceDeleteHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("123", "jhgfdvxc")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "abc", bus.Inbox[0].(model.NewHolderEvent).CurrentHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).PrevHolderUserId)
	assert.Equal(t, "jhgfdvxc", bus.Inbox[0].(model.NewHolderEvent).AuthorUserId)
}

func TestNewHolderEventSelfDeleteNotHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("abc", "abc")
	assert.Nil(t, err)
	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEventForceDeleteNotHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	err := service.DeleteById("abc", "jjfftg")
	assert.Nil(t, err)
	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEventPopAnotherUser(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	_, err := service.Pop("abc")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "abc", bus.Inbox[0].(model.NewHolderEvent).CurrentHolderUserId)
	assert.Equal(t, "123", bus.Inbox[0].(model.NewHolderEvent).PrevHolderUserId)
	assert.Equal(t, "abc", bus.Inbox[0].(model.NewHolderEvent).AuthorUserId)
}

func TestNewHolderEventPopOnEmpty(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	_, err := service.Pop("123")
	assert.Equal(t, usecase.QueueIsEmpty, err)
	assert.Empty(t, bus.Inbox)
}

func Test_NewHolderEvent_OnPass(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus, sync.Mutex{}, nil}

	assert.Equal(t, usecase.YouAreNotHolder, service.Pass("5653"))
	assert.Empty(t, bus.Inbox)
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "4"}))
	assert.Equal(t, usecase.NoOneToPass, service.Pass("4"))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "6"}))
	assert.Nil(t, service.Pass("4"))
	//6 4
	containsNewHolderEvent(bus.Inbox, "6", "4", "4")
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "4"})
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "1"}))
	assert.Nil(t, service.Add(model.QueueEntity{UserId: "17"}))
	//6 4 1 17
	assert.Equal(t, usecase.YouAreNotHolder, service.Pass("4"))
	assert.Nil(t, service.Pass("6"))
	//4 6 1 17
	containsNewHolderEvent(bus.Inbox, "4", "6", "6")
	assert.Contains(t, bus.Inbox, model.NewSecondEvent{CurrentSecondUserId: "6"})
}

func containsNewHolderEvent(inbox []interface{}, cur, au, prev string) bool {
	for _, e := range inbox {
		if event, ok := e.(model.NewHolderEvent); ok {
			if event.CurrentHolderUserId == cur && event.AuthorUserId == au && event.PrevHolderUserId != prev {
				return true
			}
		}
	}
	return false
}
