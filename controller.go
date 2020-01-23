package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/yonesko/slack-queue-bot/action"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/user"
	"log"
	"strings"
)

type Controller struct {
	rtm             *slack.RTM
	api             *slack.Client
	queueService    action.Service
	userInfoCache   map[string]model.User
	logger          *log.Logger
	userRepository  user.Repository
	queueRepository queue.Repository
}

func newController(slackApi *slack.Client, userRepository user.Repository, queueRepository queue.Repository) *Controller {
	rtm := slackApi.NewRTM()
	go rtm.ManageConnection()
	return &Controller{
		rtm:             rtm,
		api:             slackApi,
		queueService:    action.NewService(),
		userInfoCache:   map[string]model.User{},
		logger:          log.New(lumberWriter, "controller: ", log.Lshortfile|log.LstdFlags),
		userRepository:  userRepository,
		queueRepository: queueRepository,
	}
}

func (cont *Controller) handleMessageEvent(ev *slack.MessageEvent) {
	defer func() {
		if r := recover(); r != nil {
			cont.logger.Printf("catch panic: %#v", r)
		}
	}()
	cont.logger.Printf("process event: %#v", ev)
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
	err := cont.queueService.Add(model.QueueEntity{UserId: ev.User})
	if err == action.AlreadyExistErr {
		txt := i18n.P.MustGetString("you_are_already_in_the_queue")
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(txt, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
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
	txt := i18n.P.MustGetString("error_occurred")
	cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(txt, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
}

func (cont *Controller) findHolder() (*model.QueueEntity, error) {
	q, err := cont.queueService.Show()
	if err != nil {
		return nil, err
	}
	if len(q.Entities) == 0 {
		return nil, nil
	}
	return &q.Entities[0], nil
}

func (cont *Controller) deleteUser(ev *slack.MessageEvent) {
	holder, err := cont.findHolder()
	if err != nil {
		cont.reportError(ev)
		cont.logger.Print(err)
		return
	}
	deletedEntity := model.QueueEntity{UserId: ev.User}
	switch cont.queueService.Delete(deletedEntity) {
	case action.NoSuchUserErr:
		txt := i18n.P.MustGetString("you_are_not_in_the_queue")
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(txt, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
		cont.showQueue(ev)
	case nil:
		if holder != nil && deletedEntity == *holder {
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
	if len(q.Entities) > 0 {
		firstUser := q.Entities[0]
		user, err := cont.userRepository.FindById(firstUser.UserId)
		if err != nil {
			cont.reportError(ev)
			cont.logger.Print(err)
			return
		}
		txt := fmt.Sprintf(i18n.P.MustGetString("your_turn_came"), user.DisplayName)
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
	if len(q.Entities) == 0 {
		txt := i18n.P.MustGetString("queue_is_empty")
		cont.rtm.SendMessage(cont.rtm.NewOutgoingMessage(txt, ev.Channel, slack.RTMsgOptionTS(ev.ThreadTimestamp)))
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

func (cont *Controller) composeShowQueueText(queue model.Queue, userId string) (string, error) {
	txt := ""
	for i, u := range queue.Entities {
		user, err := cont.userRepository.FindById(u.UserId)
		if err != nil {
			return "", fmt.Errorf("can't composeShowQueueText: %s", err)
		}
		highlight := ""
		if u.UserId == userId {
			highlight = ":point_left::skin-tone-2:"
		}
		txt += fmt.Sprintf("`%dº` %s (%s) %s\n", i+1, user.FullName, user.DisplayName, highlight)
	}
	return txt, nil
}

func (cont *Controller) showHelp(ev *slack.MessageEvent) {
	txt := fmt.Sprintf(i18n.P.MustGetString("help_text"), cont.title(ev))
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

func (cont *Controller) title(ev *slack.MessageEvent) string {
	if user, err := cont.userRepository.FindById(ev.User); err == nil {
		return strings.TrimSpace(user.FullName)
	}
	return "human"
}
