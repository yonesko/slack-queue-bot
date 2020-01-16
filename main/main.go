package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

func main() {
	srv := NewServer()

	for msg := range srv.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			switch strings.TrimSpace(ev.Text) {
			case "add":
				srv.handlerAdd(ev)
			case "del":
				srv.handlerDel(ev)
			case "show":
				srv.handlerShow(ev)
			default:
				srv.handlerHelp(ev)
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
		return "", fmt.Errorf("env var " + name + " is absent today")
	}
	return s, nil
}
