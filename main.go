package main

import (
	"fmt"
	"log"
	"os"
	"slack-queue-bot/queue"
	"strings"

	"github.com/nlopes/slack"
)

func main() {
	env, err := getenv("SLACK_QUEUE_BOT_TOKEN")
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

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			switch {
			case strings.HasPrefix(ev.Text, "add"):
				handlerAdd(queueService, ev, rtm)
			case strings.HasPrefix(ev.Text, "del"):
				err := queueService.Delete(queue.User{Id: ev.User})
				if err == queue.NoSuchUser {
					rtm.SendMessage(rtm.NewOutgoingMessage("You are not in the queue", ev.Channel))
					break
				}
				if err != nil {
					rtm.SendMessage(rtm.NewOutgoingMessage("Some error occurred :(", ev.Channel))
					break
				}
				q, err := queueService.Show()
				if err != nil {
					rtm.SendMessage(rtm.NewOutgoingMessage("Some error occurred :(", ev.Channel))
					break
				}
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
			case strings.HasPrefix(ev.Text, "show"):
				q, err := queueService.Show()
				if err != nil {
					rtm.SendMessage(rtm.NewOutgoingMessage("Some error occurred:(", ev.Channel))
					break
				}
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
			}
		case *slack.OutgoingErrorEvent:
			fmt.Printf("Can't send msg: %s", ev.Error())
		case *slack.InvalidAuthEvent, *slack.ConnectionErrorEvent:
			log.Fatal(msg)
		}
	}
}

func handlerAdd(queueService queue.Service, ev *slack.MessageEvent, rtm *slack.RTM) {
	err := queueService.Add(queue.User{Id: ev.User, Channel: ev.User})
	if err == queue.AlreadyExistErr {
		rtm.SendMessage(rtm.NewOutgoingMessage("You are already in the queue", ev.Channel))
		return
	}
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage("Some error occurred :(", ev.Channel))
		return
	}
	q, err := queueService.Show()
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage("Some error occurred :(", ev.Channel))
		return
	}
	rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(q), ev.Channel))
}

func getenv(name string) (string, error) {
	s := os.Getenv(name)
	if len(s) == 0 {
		return "", fmt.Errorf("env var " + name + " is absent today")
	}
	return s, nil
}
