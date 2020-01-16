package main

import (
	errors "dwatcher/pkg/dep/sources/https---github.com-pkg-errors"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
	"slack-queue-bot/queue"
	"strings"
)

type Server struct {
	rtm           *slack.RTM
	api           *slack.Client
	queueService  queue.Service
	userInfoCache map[string]*slack.User
}

func NewServer() *Server {
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
	return &Server{
		rtm:          rtm,
		api:          api,
		queueService: queueService,
	}
}

const unexpectedErrorText = "Some error has occurred :("

func (s *Server) addUser(ev *slack.MessageEvent) {
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

func (s *Server) deleteUser(ev *slack.MessageEvent) {
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

func (s *Server) showQueue(ev *slack.MessageEvent) {
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

func (s *Server) composeShowText(queue queue.Queue) (string, error) {
	txt := ""
	for i, u := range queue.Users {
		info, err := s.getUserInfo(u.Id)
		if err != nil {
			return "", errors.WithMessage(err, "can't composeShowText")
		}
		txt += fmt.Sprintf(">`%dº` %s (%s)\n", i+1, info.RealName, info.Name)
	}
	return txt, nil
}

func (s *Server) getUserInfo(userId string) (*slack.User, error) {
	info, err := s.api.GetUserInfo(userId)
	if info, exists := s.userInfoCache[userId]; err != nil && exists {
		return info, nil
	}
	return info, err
}

func (s *Server) showHelp(ev *slack.MessageEvent) {
	template := "Hello %s, This is my API:\n" +
		"`add` - Add you to the queue\n" +
		"`del` - Delete you of the queue\n" +
		"`show` - Show the queue\n"
	txt := fmt.Sprintf(template, title(s, ev))
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(txt, ev.Channel))
}

func title(s *Server, ev *slack.MessageEvent) string {
	title := "human"
	info, err := s.getUserInfo(ev.User)
	if err == nil {
		title = strings.TrimSpace(info.RealName)
	}
	return title
}
