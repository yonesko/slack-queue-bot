package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
	"slack-queue-bot/queue"
	"strings"
)

type Controller struct {
	rtm           *slack.RTM
	api           *slack.Client
	queueService  queue.Service
	userInfoCache map[string]*slack.User
}

func NewController() *Controller {
	api := slack.New(
		mustGetEnv("BOT_USER_OAUTH_ACCESS_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()
	return &Controller{
		rtm:           rtm,
		api:           api,
		queueService:  queue.NewService(),
		userInfoCache: map[string]*slack.User{},
	}
}

const unexpectedErrorText = "Some error has occurred :pepe_sad:"

func (s *Controller) addUser(ev *slack.MessageEvent) {
	err := s.queueService.Add(queue.User{Id: ev.User, Channel: ev.User})
	if err == queue.AlreadyExistErr {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage("You are already in the queue", ev.Channel))
		s.showQueue(ev)
		return
	}
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.showQueue(ev)
}

func (s *Controller) deleteUser(ev *slack.MessageEvent) {
	err := s.queueService.Delete(queue.User{Id: ev.User})
	if err == queue.NoSuchUserErr {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage("You are not in the queue", ev.Channel))
		s.showQueue(ev)
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
	if len(q.Users) > 0 {
		firstUser := q.Users[0]
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(fmt.Sprintf("<@%s> is your turn! When you finish, you should delete you from the queue", firstUser.Id), ev.Channel))
	}
}

func (s *Controller) showQueue(ev *slack.MessageEvent) {
	q, err := s.queueService.Show()
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	text, err := s.composeShowQueueText(q, ev.User)
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(text, ev.Channel))
}

func (s *Controller) composeShowQueueText(queue queue.Queue, userId string) (string, error) {
	txt := ""
	if len(queue.Users) == 0 {
		return "Queue is empty", nil
	}
	for i, u := range queue.Users {
		info, err := s.getUserInfo(u.Id)
		if err != nil {
			return "", fmt.Errorf("can't composeShowQueueText: %s", err)
		}
		highlight := ""
		if u.Id == userId {
			highlight = ":point_left::skin-tone-2:"
		}
		txt += fmt.Sprintf("`%dº` %s (%s) %s\n", i+1, info.RealName, info.Name, highlight)
	}
	return txt, nil
}

func (s *Controller) getUserInfo(userId string) (*slack.User, error) {
	if info, exists := s.userInfoCache[userId]; exists {
		return info, nil
	}
	info, err := s.api.GetUserInfo(userId)
	if err != nil {
		return nil, err
	}
	s.userInfoCache[userId] = info
	return info, nil
}

func (s *Controller) showHelp(ev *slack.MessageEvent) {
	template := "Hello, %s, This is my API:\n" +
		"`add` - Add you to the queue\n" +
		"`del` - Delete you of the queue\n" +
		"`show` - Show the queue\n" +
		"`clean` - Clean all\n" +
		"`pop` - Delete first user of the queue\n"
	txt := fmt.Sprintf(template, title(s, ev))
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(txt, ev.Channel))
}

func (s *Controller) clean(ev *slack.MessageEvent) {
	err := s.queueService.DeleteAll()
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.showQueue(ev)
}

func (s *Controller) pop(ev *slack.MessageEvent) {
	err := s.queueService.Pop()
	if err != nil {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.showQueue(ev)
}

func title(s *Controller, ev *slack.MessageEvent) string {
	title := "human"
	info, err := s.getUserInfo(ev.User)
	if err == nil {
		title = strings.TrimSpace(info.RealName)
	}
	return title
}
