package main

import (
	errors "dwatcher/pkg/dep/sources/https---github.com-pkg-errors"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
	"slack-queue-bot/queue"
)

type Server struct {
	rtm          *slack.RTM
	api          *slack.Client
	queueService queue.Service
}

func NewServer() Server {
	env, err := getenv("BOT_USER_OAUTH_ACCESS_TOKEN")
	if err != nil {
		log.Fatal(err)
	}
	api := slack.New(
		env,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	queueService := queue.NewService()
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	return Server{
		rtm:          rtm,
		api:          api,
		queueService: queueService,
	}
}

const unexpectedErrorText = "Some error has occurred :("

func (s Server) handlerAdd(ev *slack.MessageEvent) {
	err := s.queueService.Add(queue.User{Id: ev.User, Channel: ev.User})
	if err == queue.AlreadyExistErr {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage("You are already in the queue", ev.Channel))
		return
	}
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	q, err := s.queueService.Show()
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
}

func (s Server) handlerDel(ev *slack.MessageEvent) {
	err := s.queueService.Delete(queue.User{Id: ev.User})
	if err == queue.NoSuchUser {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage("You are not in the queue", ev.Channel))
		return
	}
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	q, err := s.queueService.Show()
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
}

func (s Server) handlerShow(ev *slack.MessageEvent) {
	q, err := s.queueService.Show()
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	text, err := s.composeShowText(q)
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(text, ev.Channel))
}

func (s Server) composeShowText(queue queue.Queue) (string, error) {
	txt := ""
	for i, u := range queue.Users {
		info, err := s.api.GetUserInfo(u.Id)
		if err != nil {
			return "", errors.WithMessage(err, "can't composeShowText")
		}
		txt += fmt.Sprintf("%d %s (%s)\n", i+1, info.RealName, info.Name)
	}
	return txt, nil
}
