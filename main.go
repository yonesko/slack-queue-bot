package main

import (
	"fmt"
	"github.com/yonesko/slack-queue-bot/controller"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

const thisBotUserId = "<@USMRFHHPE>"

func main() {
	ctrl := controller.NewController(mustGetEnv("BOT_USER_OAUTH_ACCESS_TOKEN"))

	for msg := range ctrl.Rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			switch extractCommand(ev.Text) {
			case "add":
				ctrl.AddUser(ev)
			case "del":
				ctrl.DeleteUser(ev)
			case "show":
				ctrl.ShowQueue(ev)
			case "clean":
				ctrl.Clean(ev)
			case "pop":
				ctrl.Pop(ev)
			default:
				ctrl.ShowHelp(ev)
			}
		case *slack.OutgoingErrorEvent:
			fmt.Printf("Can't send msg: %s", ev.Error())
		case *slack.InvalidAuthEvent, *slack.ConnectionErrorEvent:
			log.Fatal(msg)
		}
	}
}

func needProcess(m *slack.MessageEvent) bool {
	mention := strings.HasPrefix(m.Text, thisBotUserId)
	isDirect := strings.HasPrefix(m.Channel, "D")
	simple := m.SubType == "" && !m.Hidden
	return simple && (isDirect || mention)
}

func extractCommand(text string) string {
	txt := strings.Replace(text, thisBotUserId, "", 1)
	txt = strings.ToLower(txt)
	return strings.TrimSpace(txt)
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}
