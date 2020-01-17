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
	srv := NewServer()

	for msg := range srv.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			switch extractCommand(ev.Text) {
			case "add":
				srv.addUser(ev)
			case "del":
				srv.deleteUser(ev)
			case "show":
				srv.showQueue(ev)
			case "clean":
				srv.clean(ev)
			case "pop":
				srv.pop(ev)
			default:
				srv.showHelp(ev)
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
