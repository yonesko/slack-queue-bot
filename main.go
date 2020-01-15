package main

import (
	"log"
	"os"
	"slack-queue-bot/queue"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New(
		os.Getenv("SLACK_QUEUE_BOT_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	queueManager := queue.NewService()

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			rtm.SendMessage(rtm.NewOutgoingMessage(ev.Text, ev.Channel))
			queueManager.Add(queue.User{Id: "4545"})
		}
	}
}
