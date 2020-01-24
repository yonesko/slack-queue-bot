package main

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
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

func newController(userRepository user.Repository, queueService usecase.QueueService) *Controller {
	return &Controller{
		queueService:   queueService,
		logger:         log.New(lumberWriter, "controller: ", log.Lshortfile|log.LstdFlags),
		userRepository: userRepository,
	}
}

func (cont *Controller) execute(command usecase.Command) string {
	defer func() {
		if r := recover(); r != nil {
			cont.logger.Printf("catch panic: %#v", r)
		}
	}()

	var txt string
	var err error
	switch command.Data.(type) {
	case usecase.AddCommand:
		txt, err = cont.addUser(command.AuthorUserId)
	case usecase.DelCommand:
		txt, err = cont.deleteUser(command.AuthorUserId)
	case usecase.ShowCommand:
		txt, err = cont.showQueue(command.AuthorUserId)
	case usecase.CleanCommand:
		txt, err = cont.clean(command.AuthorUserId)
	case usecase.PopCommand:
		txt, err = cont.pop(command.AuthorUserId)
	default:
		cont.logger.Printf("undefined command : %v", command)
		return cont.showHelp(command.AuthorUserId)
	}
	if err != nil {
		cont.logger.Println(err)
		return i18n.P.MustGetString("error_occurred")
	}
	return txt
}

func (cont *Controller) addUser(authorUserId string) (string, error) {
	err := cont.queueService.Add(model.QueueEntity{UserId: authorUserId})
	if err == usecase.AlreadyExistErr {
		return cont.appendQueue(i18n.P.MustGetString("you_are_already_in_the_queue"), authorUserId), nil
	}
	if err != nil {
		return "", err
	}
	return cont.appendQueue(i18n.P.MustGetString("added_successfully"), authorUserId), nil
}

func (cont *Controller) deleteUser(authorUserId string) (string, error) {
	err := cont.queueService.DeleteById(authorUserId)
	if err == usecase.NoSuchUserErr {
		return cont.appendQueue(i18n.P.MustGetString("you_are_not_in_the_queue"), authorUserId), nil
	}
	if err == usecase.QueueIsEmpty {
		return cont.showQueue(authorUserId)
	}
	if err != nil {
		return "", err
	}
	return cont.appendQueue(i18n.P.MustGetString("deleted_successfully"), authorUserId), nil
}

func (cont *Controller) appendQueue(txt string, authorUserId string) string {
	queueTxt, err := cont.showQueue(authorUserId)
	if err != nil {
		return txt
	}
	return txt + "\n" + queueTxt
}

func (cont *Controller) showQueue(authorUserId string) (string, error) {
	q, err := cont.queueService.Show()
	if err != nil {
		return "", err
	}
	text, err := cont.composeShowQueueText(q, authorUserId)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (cont *Controller) composeShowQueueText(queue model.Queue, authorUserId string) (string, error) {
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
		if u.UserId == authorUserId {
			highlight = ":point_left::skin-tone-2:"
		}
		txt += fmt.Sprintf("`%dº` %s (%s) %s\n", i+1, user.FullName, user.DisplayName, highlight)
	}
	return txt, nil
}

func (cont *Controller) showHelp(authorUserId string) string {
	return fmt.Sprintf(i18n.P.MustGetString("help_text"), cont.title(authorUserId))
}

func (cont *Controller) clean(authorUserId string) (string, error) {
	err := cont.queueService.DeleteAll()
	if err == usecase.QueueIsEmpty {
		return cont.showQueue(authorUserId)
	}
	if err != nil {
		return "", err
	}
	return cont.appendQueue(i18n.P.MustGetString("cleaned_successfully"), authorUserId), nil
}

func (cont *Controller) pop(authorUserId string) (string, error) {
	deletedUserId, err := cont.queueService.Pop()
	if err == usecase.QueueIsEmpty {
		return cont.showQueue(authorUserId)
	}
	if err != nil {
		return "", err
	}
	txt := fmt.Sprintf(i18n.P.MustGetString("popped_successfully"), cont.deletedUserTxt(deletedUserId))
	return cont.appendQueue(txt, authorUserId), nil
}

func (cont *Controller) deletedUserTxt(deletedUserId string) string {
	if user, err := cont.userRepository.FindById(deletedUserId); err == nil {
		return user.FullName
	}
	return ""
}

func (cont *Controller) title(userId string) string {
	if user, err := cont.userRepository.FindById(userId); err == nil {
		return strings.TrimSpace(user.FullName)
	}
	return "human"
}
