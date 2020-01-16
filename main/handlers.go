package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"slack-queue-bot/queue"
)

const unexpectedErrorText = "Some error has occurred :("

func handlerAdd(queueService queue.Service, ev *slack.MessageEvent, rtm *slack.RTM) {
	err := queueService.Add(queue.User{Id: ev.User, Channel: ev.User})
	if err == queue.AlreadyExistErr {
		rtm.SendMessage(rtm.NewOutgoingMessage("You are already in the queue", ev.Channel))
		return
	}
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	q, err := queueService.Show()
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
}

func handlerDel(queueService queue.Service, ev *slack.MessageEvent, rtm *slack.RTM) {
	err := queueService.Delete(queue.User{Id: ev.User})
	if err == queue.NoSuchUser {
		rtm.SendMessage(rtm.NewOutgoingMessage("You are not in the queue", ev.Channel))
		return
	}
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	q, err := queueService.Show()
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
}

func handlerShow(queueService queue.Service, ev *slack.MessageEvent, rtm *slack.RTM) {
	q, err := queueService.Show()
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
}
