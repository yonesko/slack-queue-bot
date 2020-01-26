package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/event"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/model"
	queuemock "github.com/yonesko/slack-queue-bot/queue/mock"
	"testing"
	"time"
)

func TestNewHolderEventAddToEmptyQueue(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus}

	err := service.Add(model.QueueEntity{UserId: "123"})
	time.Sleep(time.Millisecond * 5) //wait async sending
	assert.Nil(t, err)

	assert.Equal(t,
		[]interface{}{event.NewHolderEvent{
			CurrentHolderUserId: "123",
			PrevHolderUserId:    "",
			AuthorUserId:        "123",
		}},
		bus.Inbox,
	)
}
func TestNewHolderEventNotEmpty(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("abc", "abc")
	time.Sleep(time.Millisecond * 5) //wait async sending
	assert.Nil(t, err)

	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEventPopAnotherUser(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("abc")
	time.Sleep(time.Millisecond * 5) //wait async sending
	assert.Nil(t, err)

	assert.Equal(t,
		[]interface{}{event.NewHolderEvent{
			CurrentHolderUserId: "abc",
			PrevHolderUserId:    "123",
			AuthorUserId:        "abc",
		}},
		bus.Inbox,
	)
}
func TestNewHolderEventPopYourself(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}}}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("123")
	time.Sleep(time.Millisecond * 5) //wait async sending
	assert.Nil(t, err)
	assert.Equal(t,
		[]interface{}{event.NewHolderEvent{
			CurrentHolderUserId: "",
			PrevHolderUserId:    "123",
			AuthorUserId:        "123",
		}},
		bus.Inbox,
	)
}

func TestNewHolderEventPopOnEmpty(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("123")
	time.Sleep(time.Millisecond * 5) //wait async sending
	assert.Equal(t, QueueIsEmpty, err)
	assert.Empty(t, bus.Inbox)
}
