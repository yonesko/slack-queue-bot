package listener

import (
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/user"
)

type DeletedEventListener interface {
	Fire(newHolderEvent model.DeletedEvent)
}

type notifyDeletedEventListener struct {
	slackApi       *slack.Client
	gateway        gateway.Gateway
	userRepository user.Repository
}

func NewNotifyDeletedEventListener(slackApi *slack.Client) *notifyDeletedEventListener {
	return &notifyDeletedEventListener{slackApi: slackApi}
}

func (n *notifyDeletedEventListener) Fire(ev model.DeletedEvent) {
	n.gateway.SendAndLog(ev.DeletedUserId, n.deleterTxt(ev.AuthorUserId)+" выкинул тебя  из маршрутки")
}

func (n *notifyDeletedEventListener) deleterTxt(userId string) string {
	user, err := n.userRepository.FindById(userId)
	if err != nil {
		return "кто-то"
	}
	return user.FullName
}
