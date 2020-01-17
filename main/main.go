package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

const thisBotUserId = "<@USMRFHHPE>"

func main() {
	controller := NewController()

	for msg := range controller.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			switch extractCommand(ev.Text) {
			case "add":
				controller.addUser(ev)
			case "del":
				controller.deleteUser(ev)
			case "show":
				controller.showQueue(ev)
			case "clean":
				controller.clean(ev)
			case "pop":
				controller.pop(ev)
			default:
				controller.showHelp(ev)
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
	return strings.TrimSpace(strings.Replace(text, thisBotUserId, "", 1))
}

func getenv(name string) (string, error) {
	s := os.Getenv(name)
	if len(s) == 0 {
		return "", fmt.Errorf("env var " + name + " is absent today")
	}
	return s, nil
}
