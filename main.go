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
				queueService.Add(queue.User{Id: ev.User})
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(queueService.Show()), ev.Channel))
			case strings.HasPrefix(ev.Text, "del"):
				queueService.Delete(queue.User{Id: ev.User})
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(queueService.Show()), ev.Channel))
			case strings.HasPrefix(ev.Text, "show"):
				rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprint(queueService.Show()), ev.Channel))
			}
		case *slack.OutgoingErrorEvent:
			fmt.Printf("Can't send msg: %s", ev.Error())
		case *slack.InvalidAuthEvent, *slack.ConnectionErrorEvent:
			log.Fatal(msg)
		}
	}
}

func getenv(name string) (string, error) {
	s := os.Getenv(name)
	if len(s) == 0 {
		return "", fmt.Errorf(name + " is absent today")
	}
	return s, nil
}
