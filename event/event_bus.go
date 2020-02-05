package event

import (
	"github.com/yonesko/slack-queue-bot/event/listener"
	"github.com/yonesko/slack-queue-bot/model"
	"io"
	"log"
)

type QueueChangedEventBus interface {
	Send(event interface{})
}

func NewQueueChangedEventBus(lumberWriter io.Writer, newHolderEventListeners []listener.NewHolderEventListener) QueueChangedEventBus {
	return &queueChangedEventBus{
		logger:                  log.New(lumberWriter, "event-bus: ", log.Lshortfile|log.LstdFlags),
		newHolderEventListeners: newHolderEventListeners,
	}
}

type queueChangedEventBus struct {
	logger                  *log.Logger
	newHolderEventListeners []listener.NewHolderEventListener
}

func (q *queueChangedEventBus) Send(event interface{}) {
	q.logger.Printf("received event %#v", event)
	switch event := event.(type) {
	case model.NewHolderEvent:
		for _, l := range q.newHolderEventListeners {
			go l.Fire(event)
		}
	default:
		q.logger.Printf("unknown event %v", event)
	}
}
