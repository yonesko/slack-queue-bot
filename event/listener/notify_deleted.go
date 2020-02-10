package listener

import (
	"github.com/yonesko/slack-queue-bot/gateway"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/user"
)

type DeletedEventListener interface {
	Fire(newHolderEvent model.DeletedEvent)
}

type notifyDeletedEventListener struct {
	gateway        gateway.Gateway
	userRepository user.Repository
}

func NewNotifyDeletedEventListener(gateway gateway.Gateway, userRepository user.Repository) *notifyDeletedEventListener {
	return &notifyDeletedEventListener{gateway: gateway, userRepository: userRepository}
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
