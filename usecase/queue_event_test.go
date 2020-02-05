package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/event"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/model"
	queuemock "github.com/yonesko/slack-queue-bot/queue/mock"
	"testing"
)

func TestNewHolderEventAddToEmptyQueue(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus}

	err := service.Add(model.QueueEntity{UserId: "123"})
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).CurrentHolderUserId, "123")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).PrevHolderUserId, "")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).AuthorUserId, "123")
}

func TestNewHolderEventCheckTheSecond(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}, {"z"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("123", "123")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, "z", bus.Inbox[0].(event.NewHolderEvent).SecondUserId)
}
func TestNewHolderEventSelfDeleteHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("123", "123")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).CurrentHolderUserId, "abc")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).PrevHolderUserId, "123")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).AuthorUserId, "123")
}

func TestNewHolderEventForceDeleteHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("123", "jhgfdvxc")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).CurrentHolderUserId, "abc")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).PrevHolderUserId, "123")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).AuthorUserId, "jhgfdvxc")
}

func TestNewHolderEventSelfDeleteNotHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("abc", "abc")
	assert.Nil(t, err)
	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEventForceDeleteNotHolder(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("abc", "jjfftg")
	assert.Nil(t, err)
	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEventPopAnotherUser(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("abc")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).CurrentHolderUserId, "abc")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).PrevHolderUserId, "123")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).AuthorUserId, "abc")
}
func TestNewHolderEventPopYourself(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{Entities: []model.QueueEntity{{"123"}}}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("123")
	assert.Nil(t, err)
	assert.Len(t, bus.Inbox, 1)
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).CurrentHolderUserId, "")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).PrevHolderUserId, "123")
	assert.Equal(t, bus.Inbox[0].(event.NewHolderEvent).AuthorUserId, "123")
}

func TestNewHolderEventPopOnEmpty(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("123")
	assert.Equal(t, QueueIsEmpty, err)
	assert.Empty(t, bus.Inbox)
}
