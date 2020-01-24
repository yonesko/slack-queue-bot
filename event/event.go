package event

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/user"
	"io"
	"log"
)

type NewHolderEvent struct {
	CurrentHolderUserId string
}

type QueueChangedEventBus interface {
	Send(event interface{})
}

func NewQueueChangedEventBus(slackApi *slack.Client, userRepository user.Repository, lumberWriter io.Writer) QueueChangedEventBus {
	return &queueChangedEventBus{
		slackApi:       slackApi,
		userRepository: userRepository,
		logger:         log.New(lumberWriter, "event-bus: ", log.Lshortfile|log.LstdFlags),
	}
}

type queueChangedEventBus struct {
	slackApi       *slack.Client
	userRepository user.Repository
	logger         *log.Logger
}

func (q *queueChangedEventBus) Send(event interface{}) {
	switch event := event.(type) {
	case NewHolderEvent:
		q.logger.Printf("received event %#v", event)
		q.notifyNewHolder(event.CurrentHolderUserId)
	default:
		q.logger.Printf("unknown event %v", event)
	}
}

func (q *queueChangedEventBus) notifyNewHolder(userId string) {
	_, _, err := q.slackApi.PostMessage(userId,
		slack.MsgOptionText(i18n.P.MustGetString("your_turn_came"), true),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Printf("can't notify %s", err)
	}
}
