package usecase

import (
	"github.com/stretchr/testify/assert"
	"github.com/yonesko/slack-queue-bot/event"
	eventmock "github.com/yonesko/slack-queue-bot/event/mock"
	"github.com/yonesko/slack-queue-bot/model"
	queuemock "github.com/yonesko/slack-queue-bot/queue/mock"
	"testing"
)

func TestNewHolderEvent(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("123")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, bus.Inbox, []interface{}{
		event.NewHolderEvent{CurrentHolderUserId: "abc"},
	})
}
func TestNewHolderEvent2(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	err := service.DeleteById("abc")
	if err != nil {
		t.Error(err)
	}

	assert.Empty(t, bus.Inbox)
}

func TestNewHolderEvent3(t *testing.T) {
	bus := eventmock.QueueChangedEventBus{Inbox: []interface{}{}}
	queueRepository := queuemock.QueueRepository{model.Queue{[]model.QueueEntity{{"123"}, {"abc"}}}}
	service := &service{&queueRepository, &bus}

	_, err := service.Pop()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, bus.Inbox, []interface{}{
		event.NewHolderEvent{CurrentHolderUserId: "abc"},
	})
}
