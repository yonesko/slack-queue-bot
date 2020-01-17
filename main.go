package main

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

const thisBotUserId = "<@USMRFHHPE>"

func main() {
	lumberWriter := &lumberjack.Logger{
		Filename: "slack-queue-bot.log",
		MaxSize:  500,
		Compress: true,
	}
	logger := log.New(lumberWriter, "queue-bot: ", log.Lshortfile|log.LstdFlags)
	controller := newController(lumberWriter)

	for msg := range controller.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if !needProcess(ev) {
				break
			}
			logger.Printf("process event: %T", ev)
			switch extractCommand(ev.Text) {
			case "add":
				controller.AddUser(ev)
			case "del":
				controller.DeleteUser(ev)
			case "show":
				controller.ShowQueue(ev)
			case "clean":
				controller.Clean(ev)
			case "pop":
				controller.Pop(ev)
			default:
				controller.ShowHelp(ev)
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
	simple := m.SubType == "" && !m.Hidden && m.BotID == "" && m.Edited == nil
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
