package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/queue"
	"io"
	"log"
	"strings"
)

type Controller struct {
	rtm           *slack.RTM
	api           *slack.Client
	queueService  queue.Service
	userInfoCache map[string]*slack.User
	logger        *log.Logger
}

func newController(loggerWriter io.Writer) *Controller {
	api := slack.New(
		mustGetEnv("BOT_USER_OAUTH_ACCESS_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(loggerWriter, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()
	return &Controller{
		rtm:           rtm,
		api:           api,
		queueService:  queue.NewService(),
		userInfoCache: map[string]*slack.User{},
		logger:        log.New(loggerWriter, "controller: ", log.Lshortfile|log.LstdFlags),
	}
}

func (cont *Controller) handleMessageEvent(ev *slack.MessageEvent) {
	defer func() {
		if r := recover(); r != nil {
			cont.logger.Printf("catch panic: %#v", r)
		}
	}()
	cont.logger.Printf("process event: %#v", ev)
	cont = nil
	switch extractCommand(ev.Text) {
	case "add":
		cont.addUser(ev)
	case "del":
		cont.deleteUser(ev)
	case "show":
		cont.showQueue(ev)
	case "clean":
		cont.clean(ev)
	case "pop":
		cont.pop(ev)
	default:
		cont.showHelp(ev)
	}
}

func (cont *Controller) addUser(ev *slack.MessageEvent) {
	err := cont.queueService.Add(queue.User{Id: ev.User})
	if err == queue.AlreadyExistErr {
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage("You are already in the queue", ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
		cont.showQueue(ev)
		return
	}
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	cont.showQueue(ev)
}

func (cont *Controller) reportError(ev *slack.MessageEvent) {
	cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage("Some error has occurred :pepe_sad:", ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
}

func (cont *Controller) findHolder() (*queue.User, error) {
	q, err := cont.queueService.Show()
	if err != nil {
		return nil, err
	}
	if len(q.Users) == 0 {
		return nil, nil
	}
	return &q.Users[0], nil
}

func (cont *Controller) deleteUser(ev *slack.MessageEvent) {
	holder, err := cont.findHolder()
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	deletedUser := queue.User{Id: ev.User}
	switch cont.queueService.Delete(deletedUser) {
	case queue.NoSuchUserErr:
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage("You are not in the queue", ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
		cont.showQueue(ev)
	case nil:
		if holder != nil && deletedUser.Id == holder.Id {
			cont.notifyNewHolder(ev)
		}
		cont.showQueue(ev)
	default:
		cont.reportError(ev)
		cont.logger.Print(err)
	}
}

func (cont *Controller) notifyNewHolder(ev *slack.MessageEvent) {
	q, err := cont.queueService.Show()
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	if len(q.Users) > 0 {
		firstUser := q.Users[0]
		info, err := cont.getUserInfo(firstUser.Id)
		if err != nil {
			cont.reportError(ev)
			cont.logger.Print(err)
			return
		}
		txt := fmt.Sprintf("<@%cont> is your turn! When you finish, you should delete you from the queue", info.Name)
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(txt, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
	}
}

func (cont *Controller) showQueue(ev *slack.MessageEvent) {
	q, err := cont.queueService.Show()
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	if len(q.Users) == 0 {
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage("Queue is empty", ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
		return
	}
	text, err := cont.composeShowQueueText(q, ev.User)
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(text, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
}

func (cont *Controller) composeShowQueueText(queue queue.Queue, userId string) (string, error) {
	txt := ""
	for i, u := range queue.Users {
		info, err := cont.getUserInfo(u.Id)
		if err != nil {
			return "", fmt.Errorf("can't composeShowQueueText: %cont", err)
		}
		highlight := ""
		if u.Id == userId {
			highlight = ":point_left::skin-tone-2:"
		}
		txt += fmt.Sprintf("`%dº` %cont (%cont) %cont\n", i+1, info.RealName, info.Name, highlight)
	}
	return txt, nil
}

func (cont *Controller) getUserInfo(userId string) (*slack.User, error) {
	if info, exists := cont.userInfoCache[userId]; exists {
		return info, nil
	}
	info, err := cont.api.GetUserInfo(userId)
	if err != nil {
		return nil, err
	}
	cont.userInfoCache[userId] = info
	return info, nil
}

func (cont *Controller) showHelp(ev *slack.MessageEvent) {
	template := "Hello, %s, This is my API:\n" +
		"`add` - Add you to the queue\n" +
		"`del` - Delete you of the queue\n" +
		"`show` - Show the queue\n" +
		"`clean` - Clean all\n" +
		"`pop` - Delete first user of the queue\n"
	txt := fmt.Sprintf(template, title(cont, ev))
	cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(txt, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
}

func (cont *Controller) clean(ev *slack.MessageEvent) {
	err := cont.queueService.DeleteAll()
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	cont.showQueue(ev)
}

func (cont *Controller) pop(ev *slack.MessageEvent) {
	err := cont.queueService.Pop()
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	cont.notifyNewHolder(ev)
	cont.showQueue(ev)
}

func title(s *Controller, ev *slack.MessageEvent) string {
	title := "human"
	info, err := s.getUserInfo(ev.User)
	if err == nil {
		title = strings.TrimSpace(info.RealName)
	}
	return title
}
