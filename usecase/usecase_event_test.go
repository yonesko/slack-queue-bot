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

func TestNewHolderEvent(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("123", "123")
	time.Sleep(time.Millisecond * 5) //wait async sending
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t,
		[]interface{}{event.NewHolderEvent{
			CurrentHolderUserId: "abc",
			PrevHolderUserId:    "123",
			AuthorUserId:        "123",
		}},
		bus.Inbox,
	)
}
func TestNewHolderEvent2(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("abc", "")
	time.Sleep(time.Millisecond * 5) //wait async sending
	if err != nil {
		t.Error(err)
	}

	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEvent3(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop("abc")
	time.Sleep(time.Millisecond * 5) //wait async sending
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t,
		[]interface{}{event.NewHolderEvent{
			CurrentHolderUserId: "abc",
			PrevHolderUserId:    "123",
			AuthorUserId:        "abc",
		}},
		bus.Inbox,
	)
}
