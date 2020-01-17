package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/queue"
	"log"
	"os"
	"strings"
)

type Controller struct {
	Rtm           *slack.RTM
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
		Rtm:           rtm,
		api:           api,
		queueService:  queue.NewService(),
		userInfoCache: map[string]*slack.User{},
	}
}

const unexpectedErrorText = "Some error has occurred :pepe_sad:"

func (s *Controller) AddUser(ev *slack.MessageEvent) {
	err := s.queueService.Add(queue.User{Id: ev.User})
	if err == queue.AlreadyExistErr {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage("You are already in the queue", ev.Channel))
		s.ShowQueue(ev)
		return
	}
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.ShowQueue(ev)
}

func (s *Controller) DeleteUser(ev *slack.MessageEvent) {
	err := s.queueService.Delete(queue.User{Id: ev.User})
	if err == queue.NoSuchUserErr {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage("You are not in the queue", ev.Channel))
		s.ShowQueue(ev)
		return
	}
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.notifyHolder(ev.Channel)
}

func (s *Controller) notifyHolder(channelId string) {
	q, err := s.queueService.Show()
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, channelId))
		return
	}
	if len(q.Users) > 0 {
		firstUser := q.Users[0]
		info, err := s.getUserInfo(firstUser.Id)
		if err != nil {
			s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, channelId))
			return
		}
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(fmt.Sprintf("<@%s> is your turn! When you finish, you should delete you from the queue", info.Name), channelId))
	}
}

func (s *Controller) ShowQueue(ev *slack.MessageEvent) {
	q, err := s.queueService.Show()
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	text, err := s.composeShowQueueText(q, ev.User)
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(text, ev.Channel))
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

func (s *Controller) ShowHelp(ev *slack.MessageEvent) {
	template := "Hello, %s, This is my API:\n" +
		"`add` - Add you to the queue\n" +
		"`del` - Delete you of the queue\n" +
		"`show` - Show the queue\n" +
		"`clean` - Clean all\n" +
		"`pop` - Delete first user of the queue\n"
	txt := fmt.Sprintf(template, title(s, ev))
	s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(txt, ev.Channel))
}

func (s *Controller) Clean(ev *slack.MessageEvent) {
	err := s.queueService.DeleteAll()
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.ShowQueue(ev)
}

func (s *Controller) Pop(ev *slack.MessageEvent) {
	err := s.queueService.Pop()
	if err != nil {
		s.Rtm.SendMessage(s.Rtm.NewOutgoingMessage(unexpectedErrorText, ev.Channel))
		return
	}
	s.notifyHolder(ev.Channel)
}

func title(s *Controller, ev *slack.MessageEvent) string {
	title := "human"
	info, err := s.getUserInfo(ev.User)
	if err == nil {
		title = strings.TrimSpace(info.RealName)
	}
	return title
}