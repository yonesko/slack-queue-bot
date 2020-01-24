package main

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/queue"
	"github.com/yonesko/slack-queue-bot/usecase"
	"github.com/yonesko/slack-queue-bot/user"
	"log"
	"strings"
)

type Controller struct {
	queueService   usecase.QueueService
	logger         *log.Logger
	userRepository user.Repository
}

func newController(userRepository user.Repository, queueRepository queue.Repository) *Controller {
	return &Controller{
		queueService:   usecase.NewQueueService(queueRepository),
		logger:         log.New(lumberWriter, "controller: ", log.Lshortfile|log.LstdFlags),
		userRepository: userRepository,
	}
}

//return text to answer or error
func (cont *Controller) execute(command usecase.Command) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			cont.logger.Printf("catch panic: %#v", r)
		}
	}()

	switch command.Data.(type) {
	case usecase.AddCommand:
		return cont.addUser(command.AuthorUserId)
	case usecase.DelCommand:
		return cont.deleteUser(command.AuthorUserId)
	case usecase.ShowCommand:
		return cont.showQueue(command.AuthorUserId)
	case usecase.CleanCommand:
		return cont.clean(command.AuthorUserId)
	case usecase.PopCommand:
		return cont.pop(command.AuthorUserId)
	}
	cont.logger.Printf("undefined command : %v", command)
	return cont.showHelp(command.AuthorUserId), nil
}

func (cont *Controller) addUser(userId string) (string, error) {
	err := cont.queueService.Add(model.QueueEntity{UserId: userId})
	if err == usecase.AlreadyExistErr {
		txt := i18n.P.MustGetString("you_are_already_in_the_queue")
		return cont.appendQueue(txt, userId), nil
	}
	if err != nil {
		return "", err
	}
	return cont.showQueue(userId)
}

func (cont *Controller) deleteUser(userId string) (string, error) {
	err := cont.queueService.DeleteById(userId)
	if err == usecase.NoSuchUserErr {
		txt := i18n.P.MustGetString("you_are_not_in_the_queue")
		return cont.appendQueue(txt, userId), nil
	}
	if err != nil {
		return "", err
	}
	//todo add deleted successfully msg
	return cont.showQueue(userId)
}

func (cont *Controller) appendQueue(txt string, userId string) string {
	queueTxt, err := cont.showQueue(userId)
	if err != nil {
		return txt
	}
	return txt + "\n" + queueTxt
}

func (cont *Controller) showQueue(userId string) (string, error) {
	q, err := cont.queueService.Show()
	if err != nil {
		return "", err
	}
	text, err := cont.composeShowQueueText(q, userId)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (cont *Controller) composeShowQueueText(queue model.Queue, userId string) (string, error) {
	if len(queue.Entities) == 0 {
		return i18n.P.MustGetString("queue_is_empty"), nil
	}
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

func (cont *Controller) showHelp(userId string) string {
	txt := fmt.Sprintf(i18n.P.MustGetString("help_text"), cont.title(userId))
	return txt
}

func (cont *Controller) clean(userId string) (string, error) {
	err := cont.queueService.DeleteAll()
	if err != nil {
		return "", err
	}
	//todo removed all to msg
	//ignore err on showQueue
	return cont.showQueue(userId)
}

func (cont *Controller) pop(userId string) (string, error) {
	if err := cont.queueService.Pop(); err != nil {
		return "", err
	}
	//todo removed all to msg
	//ignore err on showQueue
	return cont.showQueue(userId)
}

func (cont *Controller) title(userId string) string {
	if user, err := cont.userRepository.FindById(userId); err == nil {
		return strings.TrimSpace(user.FullName)
	}
	return "human"
}
