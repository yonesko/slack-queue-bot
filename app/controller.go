package app

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/estimate"
	"github.com/yonesko/slack-queue-bot/i18n"
	"github.com/yonesko/slack-queue-bot/model"
	"github.com/yonesko/slack-queue-bot/usecase"
	"github.com/yonesko/slack-queue-bot/user"
	"io"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

type Controller struct {
	queueService       usecase.QueueService
	estimateRepository estimate.Repository
	logger             *log.Logger
	userRepository     user.Repository
}

func newController(lumberWriter io.Writer, userRepository user.Repository, queueService usecase.QueueService, estimateRepository estimate.Repository) *Controller {
	return &Controller{
		queueService:       queueService,
		logger:             log.New(lumberWriter, "controller: ", log.Lshortfile|log.LstdFlags),
		userRepository:     userRepository,
		estimateRepository: estimateRepository,
	}
}

func (c *Controller) execute(command usecase.Command) string {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Printf("catch panic: %#v", r)
			debug.PrintStack()
		}
	}()

	var txt string
	var err error
	switch command.Data.(type) {
	case usecase.AddCommand:
		txt, err = c.addUser(command.AuthorUserId)
	case usecase.DelCommand:
		txt, err = c.deleteUser(command.AuthorUserId)
	case usecase.ShowCommand:
		txt, err = c.showQueue(command.AuthorUserId)
	case usecase.CleanCommand:
		txt, err = c.clean(command.AuthorUserId)
	case usecase.PopCommand:
		txt, err = c.pop(command.AuthorUserId)
	case usecase.AckCommand:
		txt, err = c.ack(command.AuthorUserId)
	default:
		c.logger.Printf("undefined command : %v", command)
		return c.showHelp(command.AuthorUserId)
	}
	if err != nil {
		c.logger.Println(err)
		return i18n.P.MustGetString("error_occurred")
	}
	return txt
}

func (c *Controller) addUser(authorUserId string) (string, error) {
	err := c.queueService.Add(model.QueueEntity{UserId: authorUserId})
	if err == usecase.AlreadyExistErr {
		return c.appendQueue(i18n.P.MustGetString("you_are_already_in_the_queue"), authorUserId), nil
	}
	if err != nil {
		return "", err
	}
	return c.appendQueue(i18n.P.MustGetString("added_successfully"), authorUserId), nil
}

func (c *Controller) deleteUser(authorUserId string) (string, error) {
	err := c.queueService.DeleteById(authorUserId, authorUserId)
	if err == usecase.NoSuchUserErr {
		return c.appendQueue(i18n.P.MustGetString("you_are_not_in_the_queue"), authorUserId), nil
	}
	if err == usecase.QueueIsEmpty {
		return c.showQueue(authorUserId)
	}
	if err != nil {
		return "", err
	}
	return c.appendQueue(i18n.P.MustGetString("deleted_successfully"), authorUserId), nil
}

func (c *Controller) appendQueue(txt string, authorUserId string) string {
	queueTxt, err := c.showQueue(authorUserId)
	if err != nil {
		return txt
	}
	return txt + "\n" + queueTxt
}

func (c *Controller) showQueue(authorUserId string) (string, error) {
	q, err := c.queueService.Show()
	if err != nil {
		return "", err
	}
	text, err := c.composeShowQueueText(q, authorUserId)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (c *Controller) composeShowQueueText(queue model.Queue, authorUserId string) (string, error) {
	if len(queue.Entities) == 0 {
		return i18n.P.MustGetString("queue_is_empty"), nil
	}
	txt := ""
	for i, u := range queue.Entities {
		user, err := c.userRepository.FindById(u.UserId)
		if err != nil {
			return "", fmt.Errorf("can't composeShowQueueText: %s", err)
		}
		txt += fmt.Sprintf(
			"`%dº` %s (%s) %s%s%s\n",
			i+1,
			user.FullName,
			user.DisplayName,
			c.highlightTxt(u, authorUserId, i, queue),
			holdTimeTs(i, queue),
			isSleepingTxt(i, queue),
		)
	}
	return txt, nil
}

func (c *Controller) highlightTxt(u model.QueueEntity, authorUserId string, i int, queue model.Queue) string {
	if u.UserId == authorUserId {
		txt := ":point_left::skin-tone-2:"
		if i > 0 {
			txt += c.estimateTxt(i, queue)
		}
		return txt
	}
	return ""
}

func holdTimeTs(i int, queue model.Queue) string {
	if i == 0 && queue.HoldTs.Unix() > 0 {
		return " :lock: " + time.Now().Sub(queue.HoldTs).Round(time.Minute).String()
	}
	return ""
}
func isSleepingTxt(i int, queue model.Queue) string {
	if queue.HolderIsSleeping && i == 0 {
		return " :sleeping:"
	}
	return ""
}

func (c *Controller) estimateTxt(i int, queue model.Queue) string {
	estimate, err := c.estimateRepository.Read()
	if err != nil {
		c.logger.Printf("composeShowQueueText can't get estimate %s", err)
		return ""
	}
	duration := estimate.TimeToWait(uint(i), queue.HoldTs).Round(time.Minute)
	if duration == 0 {
		return ""
	}
	return fmt.Sprintf("~%s (%s)", duration, time.Now().Add(duration).Format("Mon Jan 2 15:04"))
}

func (c *Controller) showHelp(authorUserId string) string {
	return fmt.Sprintf(i18n.P.MustGetString("help_text"), c.title(authorUserId))
}

func (c *Controller) clean(authorUserId string) (string, error) {
	err := c.queueService.DeleteAll()
	if err == usecase.QueueIsEmpty {
		return c.showQueue(authorUserId)
	}
	if err != nil {
		return "", err
	}
	return c.appendQueue(i18n.P.MustGetString("cleaned_successfully"), authorUserId), nil
}
func (c *Controller) ack(authorUserId string) (string, error) {
	err := c.queueService.Ack(authorUserId)
	if err == usecase.YouAreNotHolder {
		return "ты не первый в очереди, твой ack не нужен", nil
	}
	if err == usecase.HolderIsNotSleeping {
		return "ack уже получен", nil
	}
	if err != nil {
		return "", err
	}
	return c.appendQueue(i18n.P.MustGetString("ack_is_ok"), authorUserId), nil

}
func (c *Controller) pop(authorUserId string) (string, error) {
	deletedUserId, err := c.queueService.Pop(authorUserId)
	if err == usecase.QueueIsEmpty {
		return c.showQueue(authorUserId)
	}
	if err != nil {
		return "", err
	}
	txt := fmt.Sprintf(i18n.P.MustGetString("popped_successfully"), c.deletedUserTxt(deletedUserId))
	return c.appendQueue(txt, authorUserId), nil
}

func (c *Controller) deletedUserTxt(deletedUserId string) string {
	if user, err := c.userRepository.FindById(deletedUserId); err == nil {
		return user.FullName
	}
	return ""
}

func (c *Controller) title(userId string) string {
	if user, err := c.userRepository.FindById(userId); err == nil {
		return strings.TrimSpace(user.FullName)
	}
	return "human"
}
